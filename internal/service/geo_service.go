package service

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	geoReportsKey = "geo:reports"
)

type GeoService interface {
	AddReport(reportID uuid.UUID, lat, lng float64)
	RemoveReport(reportID uuid.UUID)
	NearbyIDs(lat, lng, radiusMeters float64) ([]string, error)
	SetSession(token, userID string, ttl time.Duration)
	GetSession(token string) string
	DeleteSession(token string)
	SetUserLocation(userID string, lat, lng float64)
	NearbyUserIDs(lat, lng, radiusMeters float64) ([]string, error)
}

var _ GeoService = (*geoService)(nil)

type geoService struct {
	rdb *redis.Client
}

func NewGeoService(rdb *redis.Client) *geoService {
	return &geoService{rdb: rdb}
}

func (s *geoService) enabled() bool { return s.rdb != nil }

func (s *geoService) AddReport(reportID uuid.UUID, lat, lng float64) {
	if !s.enabled() {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := s.rdb.GeoAdd(ctx, geoReportsKey, &redis.GeoLocation{
			Name:      reportID.String(),
			Longitude: lng,
			Latitude:  lat,
		}).Err(); err != nil {
			log.Printf("[geo] GeoAdd error: %v", err)
		}
	}()
}

func (s *geoService) RemoveReport(reportID uuid.UUID) {
	if !s.enabled() {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := s.rdb.ZRem(ctx, geoReportsKey, reportID.String()).Err(); err != nil {
			log.Printf("[geo] ZRem error: %v", err)
		}
	}()
}

func (s *geoService) NearbyIDs(lat, lng, radiusMeters float64) ([]string, error) {
	if !s.enabled() {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	results, err := s.rdb.GeoSearch(ctx, geoReportsKey, &redis.GeoSearchQuery{
		Longitude:  lng,
		Latitude:   lat,
		Radius:     radiusMeters,
		RadiusUnit: "m",
		Sort:       "ASC",
		Count:      500,
	}).Result()
	if err != nil {
		return nil, err
	}
	return results, nil
}

const sessionPrefix = "session:"

func (s *geoService) SetSession(token, userID string, ttl time.Duration) {
	if !s.enabled() {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		s.rdb.Set(ctx, sessionPrefix+token, userID, ttl)
	}()
}

func (s *geoService) GetSession(token string) string {
	if !s.enabled() {
		return ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	val, err := s.rdb.Get(ctx, sessionPrefix+token).Result()
	if err != nil {
		return ""
	}
	return val
}

func (s *geoService) DeleteSession(token string) {
	if !s.enabled() {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		s.rdb.Del(ctx, sessionPrefix+token)
	}()
}

func (s *geoService) SetUserLocation(userID string, lat, lng float64) {
	if !s.enabled() {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		s.rdb.GeoAdd(ctx, "geo:users", &redis.GeoLocation{Name: userID, Longitude: lng, Latitude: lat})
		s.rdb.Expire(ctx, "geo:users", 30*time.Minute)
	}()
}

func (s *geoService) NearbyUserIDs(lat, lng, radiusMeters float64) ([]string, error) {
	if !s.enabled() {
		return nil, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return s.rdb.GeoSearch(ctx, "geo:users", &redis.GeoSearchQuery{
		Longitude: lng, Latitude: lat, Radius: radiusMeters, RadiusUnit: "m",
	}).Result()
}
