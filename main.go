package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/route"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/util"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file")
	}
	e := route.Init()
	data, err := util.Json.MarshalIndent(e.Routes(), "", "  ")
	if err != nil {
		panic(fmt.Sprint(err))
	}

	fmt.Println(string(data))
	e.Logger.Fatal(e.Start(":" + os.Getenv("APP_PORT")))
}
