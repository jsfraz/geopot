package database

import (
	"jsfraz/lucian-ssh-server/models"
	"jsfraz/lucian-ssh-server/utils"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TODO godoc

func SetupPostgres() {
	// Postgres
	connStr := "postgresql://" + os.Getenv("POSTGRES_USER") + ":" + os.Getenv("POSTGRES_PASSWORD") + "@" + os.Getenv("POSTGRES_SERVER") + ":" + os.Getenv("POSTGRES_PORT") + "/" + os.Getenv("POSTGRES_DB")
	postgres, err := gorm.Open(postgres.Open(connStr), &gorm.Config{Logger: logger.Default.LogMode(logger.Error)})
	if err != nil {
		log.Fatal(err)
	}
	// Postgres schema migration
	err = postgres.AutoMigrate(&models.Connection{})
	if err != nil {
		log.Fatal(err)
	}
	// Set Postgres in singleton
	utils.GetSingleton().Postgres = postgres
}
