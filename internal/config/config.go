package config

import (
	"fmt"
	"log"
	"os"
	"petunia/internal/model"
	"strings"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

func LoadConfig() *DatabaseConfig {
	if err := godotenv.Load(); err != nil {
		log.Println(".env not found, using default variable")
	}

	config := &DatabaseConfig{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Name:     os.Getenv("DB_NAME"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
	}

	if config.Host == "" {
		log.Fatal("DB_HOST missing")
	}
	if config.Port == "" {
		log.Fatal("DB_PORT missing")
	}
	if config.User == "" {
		log.Fatal("DB_USER missing")
	}
	if config.Password == "" {
		log.Fatal("DB_PASSWORD missing")
	}
	if config.Name == "" {
		log.Fatal("DB_NAME missing")
	}
	if config.SSLMode == "" {
		log.Fatal("DB_SSLMODE missing")
	}

	return config
}

func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode)
}

// func migrateRefreshTokens(db *gorm.DB) error {
// 	var tableExists bool
// 	if err := db.Raw(`SELECT EXISTS (
// 		SELECT 1
// 		FROM information_schema.tables
// 		WHERE table_schema = current_schema() AND table_name = 'refresh_tokens'
// 	)`).Scan(&tableExists).Error; err != nil {
// 		return err
// 	}
// 	if !tableExists {
// 		return nil
// 	}

// 	var cols []struct {
// 		ColumnName string
// 		IsNullable string
// 	}
// 	if err := db.Raw(`SELECT column_name, is_nullable
// 		FROM information_schema.columns
// 		WHERE table_name = 'refresh_tokens'
// 		ORDER BY ordinal_position`).Scan(&cols).Error; err != nil {
// 		return err
// 	}

// 	hasTokenHash := false
// 	hasToken := false
// 	for _, col := range cols {
// 		switch col.ColumnName {
// 		case "token_hash":
// 			hasTokenHash = true
// 		case "token":
// 			hasToken = true
// 		}
// 	}

// 	if !hasTokenHash {
// 		if err := db.Exec(`ALTER TABLE refresh_tokens ADD COLUMN token_hash text`).Error; err != nil {
// 			return err
// 		}
// 	}

// 	if hasToken {
// 		type legacyTokenRow struct {
// 			ID    uuid.UUID
// 			Token string
// 		}
// 		var rows []legacyTokenRow
// 		if err := db.Table("refresh_tokens").Select("id, token").Where("token_hash IS NULL AND token IS NOT NULL").Scan(&rows).Error; err != nil {
// 			return err
// 		}
// 		for _, row := range rows {
// 			sum := sha256.Sum256([]byte(row.Token))
// 			hash := hex.EncodeToString(sum[:])
// 			if err := db.Model(&model.RefreshToken{}).Where("id = ?", row.ID).Update("token_hash", hash).Error; err != nil {
// 				return err
// 			}
// 		}
// 	}

// 	var nullCount int64
// 	if err := db.Model(&model.RefreshToken{}).Where("token_hash IS NULL").Count(&nullCount).Error; err != nil {
// 		return err
// 	}
// 	if nullCount > 0 {
// 		return fmt.Errorf("refresh_tokens contains %d null token_hash values; set them before enforcing NOT NULL", nullCount)
// 	}

// 	if err := db.Exec(`ALTER TABLE refresh_tokens ALTER COLUMN token_hash SET NOT NULL`).Error; err != nil {
// 		return err
// 	}

// 	if hasToken {
// 		if err := db.Exec(`ALTER TABLE refresh_tokens DROP COLUMN token`).Error; err != nil {
// 			return err
// 		}
// 	}

// 	if err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS ux_refresh_tokens_token_hash ON refresh_tokens(token_hash)`).Error; err != nil {
// 		return err
// 	}

// 	return nil
// }

func InitDB(config *DatabaseConfig) {
	var err error
	DB, err = gorm.Open(postgres.Open(config.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("error during connection to database: %v", err)
	}

	// if err := migrateRefreshTokens(DB); err != nil {
	// 	log.Fatalf("error during refresh token migration: %v", err)
	// }

	if err := DB.AutoMigrate(
		&model.User{},
		&model.Pet{},
		&model.PasswordReset{},
		&model.EmailChange{},
		&model.RefreshToken{},
		&model.Family{},
		&model.FamilyMember{},
		&model.FamilyInvite{},
		&model.MapReport{},
		&model.ReportAbuse{},
		&model.Reminder{},
		&model.ReminderAck{},
		&model.ReminderFiredLog{},
		&model.UploadedFile{},
		&model.PushToken{},
		&model.Activity{},
		&model.ActivityPoint{},
	); err != nil {
		log.Fatalf("error during automigrate: %v", err)
	}
}

func UploadDir() string {
	if d := os.Getenv("UPLOAD_DIR"); d != "" {
		return d
	}
	return "./uploads"
}

func BaseURL() string {
	if u := os.Getenv("PUBLIC_BASE_URL"); u != "" {
		return strings.TrimRight(u, "/")
	}
	return "http://localhost:8080"
}
