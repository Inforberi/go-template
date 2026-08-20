package main

import (
	"log"

	"github.com/Inforberi/go-template/internal/app"
)

func main() {
	if err := app.New(); err != nil {
		log.Fatal(err)
	}
}
