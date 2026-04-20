package main

import (
	"flag"
	"log"

	appauth "github.com/go-park-mail-ru/2026_1_VKino/services/auth-service/internal/app/auth"
)

func main() {
	configPath := flag.String("config", "", "config file path")
	flag.Parse()

	if err := appauth.Run(*configPath); err != nil {
		log.Fatal(err)
	}
}
