package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/anshulkanwar/night-fury-radar/internal/hiccup"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	server := hiccup.NewServer()
	server.Start()

	http.HandleFunc("/cpu", server.HandleMetricRequest("cpu"))
	http.HandleFunc("/memory", server.HandleMetricRequest("memory"))

	go func() {
		http.ListenAndServe(":8080", nil)
	}()

	<-ctx.Done()
	log.Println("Shutting down gracefully...")
	server.Stop()
}
