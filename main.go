package main

import (
	"embed"
	"log"
	"os"
	"os/signal"
	"syscall"

	"mercury/app"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	if err := app.Run(assets); err != nil {
		log.Fatal(err)
	}
}

// To handle shutdown signals gracefully.
func init() {
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		os.Exit(0)
	}()
}
