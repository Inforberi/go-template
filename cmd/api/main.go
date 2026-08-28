package main

import (
	"log"

	"github.com/Inforberi/go-template/internal/app"
)

// @title			Go Template API
// @version		1.0
// @description	HTTP API generated from Go annotations.
// @BasePath		/
func main() {
	if err := app.New(); err != nil {
		log.Fatal(err)
	}
}
