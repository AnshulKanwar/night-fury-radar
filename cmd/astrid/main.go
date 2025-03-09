package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/anshulkanwar/night-fury-radar/internal/astrid/collectors"
	"github.com/anshulkanwar/night-fury-radar/internal/astrid/processor"
)

func main() {
	// Create context with cancellation on interrupt
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Initialize the processor
	processor := processor.NewMetricProcessor()
	processor.AddCollector(collectors.CpuCollector{})

	// Start collecting metrics
	log.Println("Starting metric collection...")
	processor.Start()

	// Wait for interrupt signal
	<-ctx.Done()
	log.Println("Shutting down gracefully...")
	processor.Stop()
}
