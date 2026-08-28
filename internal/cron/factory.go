package cron

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"petunia/internal/model"
	"petunia/internal/repository"
	"petunia/internal/service"
	"time"
)

type Config struct {
	Enabled              bool `json:"enabled"`
	StartAtBoot          bool `json:"start_at_boot"`
	CheckIntervalMinutes int  `json:"check_interval_minutes"`
}

func ConfigFromFile(path string) Config {
	if data, err := os.ReadFile(path); err == nil { // nolint:gosec
		var cfg Config
		if err := json.Unmarshal(data, &cfg); err == nil {
			log.Printf("[cron] loaded config from %s", path)
			return cfg
		}
		log.Printf("[cron] failed to parse %s: %v — falling back to env", path, err)
	}
	return ConfigFromEnv()
}

func ConfigFromEnv() Config {
	enabled := os.Getenv("CRON_ENABLED") != "false"
	startAtBoot := os.Getenv("CRON_START_AT_BOOT") != "false"
	interval := 1
	if v := os.Getenv("CRON_CHECK_INTERVAL_MINUTES"); v != "" {
		var n int
		if _, err := fmt.Sscanf(v, "%d", &n); err == nil && n > 0 {
			interval = n
		}
	}
	return Config{
		Enabled:              enabled,
		StartAtBoot:          startAtBoot,
		CheckIntervalMinutes: interval,
	}
}

type Factory struct {
	cfg          Config
	stop         chan struct{}
	reminderRepo repository.ReminderRepository
	reminderSvc  service.ReminderService
}

func NewFactory(
	cfg Config,
	reminderRepo repository.ReminderRepository,
	reminderSvc service.ReminderService,
) *Factory {
	return &Factory{
		cfg:          cfg,
		stop:         make(chan struct{}),
		reminderRepo: reminderRepo,
		reminderSvc:  reminderSvc,
	}
}

func (f *Factory) Start() {
	if !f.cfg.Enabled {
		log.Println("[cron] disabled via config, skipping")
		return
	}

	interval := time.Duration(f.cfg.CheckIntervalMinutes) * time.Minute

	go func() {
		if f.cfg.StartAtBoot {
			f.tick()
		}

		now := time.Now()
		waitUntil := now.Truncate(time.Minute).Add(time.Minute)
		select {
		case <-time.After(time.Until(waitUntil)):
		case <-f.stop:
			return
		}

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		f.tick()

		for {
			select {
			case <-ticker.C:
				f.tick()
			case <-f.stop:
				log.Println("[cron] stopped")
				return
			}
		}
	}()

	log.Printf("[cron] polling scheduler started (interval=%v, startAtBoot=%v)", interval, f.cfg.StartAtBoot)
}

func (f *Factory) Stop() {
	close(f.stop)
}

func (f *Factory) tick() {
	now := time.Now()
	nowMinute := now.Truncate(time.Minute)
	occurrenceKey := nowMinute.UTC().Format("2006-01-02T15:04")

	reminders, err := f.reminderRepo.FindAllEnabled()
	if err != nil {
		log.Printf("[cron] error loading reminders: %v", err)
		return
	}

	for _, rem := range reminders {
		r := rem
		if f.shouldFire(&r, nowMinute) {
			log.Printf("[cron] firing reminder %s (%s) occ=%s", r.ID, r.Title, occurrenceKey)
			go f.reminderSvc.FireReminder(&r, occurrenceKey)
		}
	}
}

func (f *Factory) shouldFire(r *model.Reminder, t time.Time) bool {
	switch r.Repeat {
	case model.ReminderRepeatDaily:
		h, m := parseTimeOfDay(r.TimeOfDay)
		return t.Hour() == h && t.Minute() == m

	case model.ReminderRepeatWeekly:
		h, m := parseTimeOfDay(r.TimeOfDay)
		dow := 0
		if r.DayOfWeek != nil {
			dow = *r.DayOfWeek
		}
		return int(t.Weekday()) == dow && t.Hour() == h && t.Minute() == m

	case model.ReminderRepeatCustom:
		return matchesCron(r.CronExpr, t)

	case model.ReminderRepeatNone:
		if r.ScheduledAt == nil {
			return false
		}
		scheduled := r.ScheduledAt.Truncate(time.Minute)
		return scheduled.Equal(t)

	default:
		return false
	}
}

func parseTimeOfDay(s string) (int, int) {
	var h, m int
	if _, err := fmt.Sscanf(s, "%d:%d", &h, &m); err != nil {
		return 8, 0
	}
	return h, m
}

func matchesCron(expr string, t time.Time) bool {
	var minF, hourF, domF, monF, dowF string
	n, err := fmt.Sscanf(expr, "%s %s %s %s %s", &minF, &hourF, &domF, &monF, &dowF)
	if err != nil || n != 5 {
		log.Printf("[cron] invalid cron expression %q: %v", expr, err)
		return false
	}

	match := func(field string, value int) bool {
		if field == "*" {
			return true
		}
		var v int
		if _, err := fmt.Sscanf(field, "%d", &v); err != nil {
			return false
		}
		return v == value
	}

	return match(minF, t.Minute()) &&
		match(hourF, t.Hour()) &&
		match(domF, t.Day()) &&
		match(monF, int(t.Month())) &&
		match(dowF, int(t.Weekday()))
}
