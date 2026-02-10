package cron

import (
	"log"
	"net"
	"net/http"
	"time"

	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/database"
)

func StartHealthChecker() {
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for range ticker.C {
			checkServices()
		}
	}()
	// Run once at start
	go checkServices()
}

func checkServices() {
	db := database.GetDB()
	if db == nil {
		log.Println("Health Check: Database not initialized")
		return
	}
	var services []database.Service
	if err := db.Find(&services).Error; err != nil {
		log.Printf("Health Check: Error fetching services: %v", err)
		return
	}

	for _, s := range services {
		status := "offline"
		if s.Protocol == "grpc" {
			if checkGRPC(s.GRPCAddr) {
				status = "online"
			}
		} else {
			if checkREST(s.BaseURL) {
				status = "online"
			}
		}

		now := time.Now()
		db.Model(&s).Updates(map[string]interface{}{
			"status":     status,
			"last_check": &now,
		})
	}
}

func checkREST(url string) bool {
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Get(url + "/health")
	if err == nil && resp.StatusCode == http.StatusOK {
		return true
	}
	// Fallback to simple GET if /health doesn't exist but service is up
	resp, err = client.Get(url)
	if err == nil {
		return true
	}
	return false
}

func checkGRPC(addr string) bool {
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
