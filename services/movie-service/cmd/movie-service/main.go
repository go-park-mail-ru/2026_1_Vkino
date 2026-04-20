package main

import (
	"flag"
	"log"

	appmovie "github.com/go-park-mail-ru/2026_1_VKino/services/movie-service/internal/app/movie"
)

func main() {
	configPath := flag.String("config", "", "config file path")
	flag.Parse()

	if err := appmovie.Run(*configPath); err != nil {
		log.Fatal(err)
	}
}
