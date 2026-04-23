package main

import (
	"chat-message-service/internal/app"
	"log"
)

func main() {
	a := app.NewApp()
	if err := a.Start(); err != nil {
		log.Fatalf("app error: %v", err)
	}
}
