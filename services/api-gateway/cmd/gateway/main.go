package main

import (
	"flag"
	"log"

	app "github.com/go-park-mail-ru/2026_1_VKino/services/api-gateway/internal/app/gateway"
)

func main() {
	configPath := flag.String("config", "", "config file path")
	flag.Parse()

	if err := app.Run(*configPath); err != nil {
		log.Fatal(err)
	}
}
