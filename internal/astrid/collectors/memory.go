package collectors

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/anshulkanwar/night-fury-radar/internal/types"
	"github.com/shirou/gopsutil/v4/mem"
)

type MemoryCollector struct{}

func (c MemoryCollector) Start(ctx context.Context, metric chan types.Metric, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case t := <-ticker.C:
				if err := c.collectAndSendMetric(t, metric); err != nil {
					log.Printf("Memory collection error: %v", err)
				}
			}
		}
	}()
}

func (c MemoryCollector) collectAndSendMetric(t time.Time, metric chan<- types.Metric) error {
	memStats, err := mem.VirtualMemory()
	if err != nil {
		return err
	}

	values := map[string]any{
		"total":       memStats.Total,
		"available":   memStats.Available,
		"used":        memStats.Used,
		"usedPercent": memStats.UsedPercent,
		"free":        memStats.Free,
	}

	select {
	case metric <- types.Metric{
		Type:      "memory",
		Values:    values,
		Timestamp: t,
	}:
		return nil
	default:
		return fmt.Errorf("metric channel full")
	}
}
