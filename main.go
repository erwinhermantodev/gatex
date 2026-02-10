package main

import (
	"fmt"

	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/config"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/cron"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/database"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/route"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/util"
)

func main() {
	cfg := config.Load()

	database.Init()
	cron.StartHealthChecker()

	e := route.Init()
	data, err := util.Json.MarshalIndent(e.Routes(), "", "  ")
	if err != nil {
		panic(fmt.Sprint(err))
	}

	fmt.Println(string(data))
	e.Logger.Fatal(e.Start(":" + cfg.AppPort))
}
