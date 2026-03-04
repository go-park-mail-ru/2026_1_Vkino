package main

import (
	"log"

	"github.com/go-park-mail-ru/2026_1_VKino/internal/app/auth"
)

func main() {
	err := auth.Run()
	if err != nil {
		log.Fatal(err)
	}
}
