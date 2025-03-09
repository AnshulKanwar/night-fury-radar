package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/anshulkanwar/night-fury-radar/internal/astrid/collectors"
	"github.com/anshulkanwar/night-fury-radar/internal/astrid/processor"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Create context with cancellation on interrupt
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Initialize the processor
	processor := processor.NewMetricProcessor()
	processor.AddCollector(collectors.CpuCollector{})
	processor.AddCollector(collectors.MemoryCollector{})

	// Start collecting metrics
	log.Println("Starting metric collection...")
	processor.Start()

	// Wait for interrupt signal
	<-ctx.Done()
	log.Println("Shutting down gracefully...")
	processor.Stop()
}
