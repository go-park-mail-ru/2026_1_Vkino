package main

import (
	"flag"
	"log"
)

func main() {
	configPath := flag.String("config", "", "config file path")

	flag.Parse()

	err := Run(configPath)
	if err != nil {
		log.Fatal(err)
	}
}
