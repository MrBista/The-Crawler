package conf

import (
	"fmt"
	"log"
	"time"

	"github.com/MrBista/The-Crawler/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(cfg DBConfig) (*gorm.DB, error) {

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s", cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port, cfg.SSLMode)
	log.Printf("db config connect dsn host=%v, user=%v, port=%v, sslmode=%v", cfg.Host, cfg.User, cfg.Port, cfg.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt: true,
		// Logger:      logger.Default.LogMode(logger.Info), // tambahkan ini
	})

	if err != nil {
		return nil, err
	}

	sqlDb, err := db.DB()
	if err != nil {
		log.Printf("Failed to connect db")
		return nil, err
	}

	if err := sqlDb.Ping(); err != nil {
		log.Printf("Failed to connect databases")
		return nil, err
	}

	if err := db.AutoMigrate(&models.CrawlPage{}); err != nil {
		log.Printf("Failed to migrate CrawlPage: %v", err)
		return nil, err
	}
	log.Println("Migration successful!")
	sqlDb.SetMaxIdleConns(10)
	sqlDb.SetMaxOpenConns(100)
	sqlDb.SetConnMaxLifetime(time.Hour)

	return db, nil
}
