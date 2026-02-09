package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/database"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/route"
)

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// Initialize DB
	db := database.Init()

	// 1. Create Services
	authService := database.Service{
		Name:     "auth-service",
		BaseURL:  os.Getenv("AUTH_SERVICE_BASE_URL"),
		Protocol: "rest", // Will be used for v1
		GRPCAddr: os.Getenv("AUTH_SERVICE_GRPC_ADDR"),
	}
	db.FirstOrCreate(&authService, database.Service{Name: "auth-service"})

	// 2. Load and Migrate Routes from auth.json
	byteFile, err := ioutil.ReadFile("route/gate/auth.json")
	if err != nil {
		log.Fatalf("Failed to read auth.json: %v", err)
	}

	var routes []route.Route
	if err := json.Unmarshal(byteFile, &routes); err != nil {
		log.Fatalf("Failed to unmarshal auth.json: %v", err)
	}

	for _, r := range routes {
		mw, _ := json.Marshal(r.Middleware)
		dbRoute := database.Route{
			Path:           r.Path,
			Method:         r.Method,
			ServiceID:      authService.ID,
			EndpointFilter: r.Endpoint, // Note: In JSON it's "endpoint_filter" but the struct tag maps it
			Tag:            r.Tag,
			Middleware:     string(mw),
		}

		// Map endpoint_filter correctly from the JSON struct if needed
		// The JSON has "endpoint_filter", but route.Route struct has "Endpoint" field with `json:"endpoint_filter"` tag

		db.FirstOrCreate(&dbRoute, database.Route{Path: r.Path})
	}

	// 3. Create Proto Mappings
	protoMappings := []database.ProtoMapping{
		{ServiceID: authService.ID, RPCMethod: "Login", ProtoPackage: "auth", RequestType: "LoginRequest", ResponseType: "LoginResponse"},
		{ServiceID: authService.ID, RPCMethod: "CheckPhone", ProtoPackage: "auth", RequestType: "CheckPhoneRequest", ResponseType: "CheckPhoneResponse"},
		{ServiceID: authService.ID, RPCMethod: "RefreshToken", ProtoPackage: "auth", RequestType: "RefreshTokenRequest", ResponseType: "RefreshTokenResponse"},
		{ServiceID: authService.ID, RPCMethod: "Logout", ProtoPackage: "auth", RequestType: "LogoutRequest", ResponseType: "StandardResponse"},
	}

	for _, pm := range protoMappings {
		db.FirstOrCreate(&pm, database.ProtoMapping{ServiceID: pm.ServiceID, RPCMethod: pm.RPCMethod})
	}

	log.Println("Migration completed successfully!")
}
