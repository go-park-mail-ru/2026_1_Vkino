package main

import (
	"flag"
	"log"
)

func main() {
	configPath := flag.String("config", "", "config file path")

	flag.Parse()

	if err := Run(*configPath); err != nil {
		log.Fatal(err)
	}
}
