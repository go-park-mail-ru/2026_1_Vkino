package main

import (
	"flag"
	"log"

	appuser "github.com/go-park-mail-ru/2026_1_VKino/services/user-service/internal/app/user"
)

func main() {
	configPath := flag.String("config", "", "config file path")
	flag.Parse()

	if err := appuser.Run(*configPath); err != nil {
		log.Fatal(err)
	}
}
