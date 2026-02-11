package database

import (
	"time"

	"gorm.io/gorm"
)

// Service represents a downstream service (e.g., auth-service)
type Service struct {
	gorm.Model
	Name      string `gorm:"uniqueIndex"`
	BaseURL   string
	Protocol  string // "rest" or "grpc"
	GRPCAddr  string
	Status    string // "online", "offline", "unknown"
	LastCheck *time.Time
}

// Route represents a gateway route mapping
type Route struct {
	gorm.Model
	Path           string `gorm:"uniqueIndex"`
	Method         string
	ServiceID      uint
	Service        Service `gorm:"foreignKey:ServiceID"`
	EndpointFilter string  // The handler identifier
	Tag            string
	Middleware     string // JSON encoded array of middleware names
}

// ProtoMapping defines the mapping for gRPC calls
type ProtoMapping struct {
	gorm.Model
	ServiceID    uint
	Service      Service `gorm:"foreignKey:ServiceID"`
	RPCMethod    string
	ServiceName  string
	ProtoPackage string
	RequestType  string
	ResponseType string
}

// ActivityLog tracks administrative actions
type ActivityLog struct {
	gorm.Model
	Action   string
	Resource string
	User     string
	Message  string
}

// RequestLog tracks all traffic through the gateway
type RequestLog struct {
	gorm.Model
	RequestID    string `gorm:"index"`
	Method       string `gorm:"index"`
	Path         string `gorm:"index"`
	StatusCode   int
	LatencyMS    int64
	ClientIP     string
	UserAgent    string
	ErrorMessage string
}

// TraceLog captures granular events for a specific request
type TraceLog struct {
	gorm.Model
	RequestID string `gorm:"index"`
	Level     string
	Component string
	Message   string
}
