package models

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	Host     string
	User     string
	Port     string
	Password string
	DBName   string
	SSLMode  string
}

var DB *gorm.DB

func InitDB(cfg Config) {

	dsn := fmt.Sprintf("host=%s user=%s port=%s password=%s dbname=%s sslmode=%s", cfg.Host, cfg.User, cfg.Port, cfg.Password, cfg.DBName, cfg.SSLMode)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&User{}); err != nil {
		panic(err)
	}


	log.Info("Database Migration successful")

	DB = db
}
