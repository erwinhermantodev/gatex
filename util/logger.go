package util

import (
	"fmt"
	"log"

	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/database"
)

func LogActivity(action, resource, user, message string) {
	db := database.GetDB()
	activity := &database.ActivityLog{
		Action:   action,
		Resource: resource,
		User:     user,
		Message:  message,
	}

	if err := db.Create(activity).Error; err != nil {
		log.Printf("Error recording activity: %v", err)
	}
}

func LogCreate(resource, user, details string) {
	LogActivity("CREATE", resource, user, fmt.Sprintf("Created new %s: %s", resource, details))
}

func LogUpdate(resource, user, details string) {
	LogActivity("UPDATE", resource, user, fmt.Sprintf("Updated %s: %s", resource, details))
}

func LogDelete(resource, user, details string) {
	LogActivity("DELETE", resource, user, fmt.Sprintf("Deleted %s: %s", resource, details))
}
