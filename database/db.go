package database

import (
	"fmt"
	"log"
	"sync"

	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	db   *gorm.DB
	once sync.Once
)

// Init initializes the database connection
func Init() *gorm.DB {
	once.Do(func() {
		cfg := config.Load()
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
			cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)

		var err error
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}

		// Auto-migrate the schema
		err = db.AutoMigrate(&Service{}, &Route{}, &ProtoMapping{}, &ActivityLog{}, &RequestLog{}, &TraceLog{})
		if err != nil {
			log.Fatalf("Failed to auto-migrate schema: %v", err)
		}
	})
	return db
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return db
}
