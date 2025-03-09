package collectors

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/anshulkanwar/night-fury-radar/internal/types"
	"github.com/shirou/gopsutil/v4/cpu"
)

type CpuCollector struct{}

func (c CpuCollector) Start(ctx context.Context, metric chan types.Metric, wg *sync.WaitGroup) {
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
					log.Printf("CPU collection error: %v", err)
				}
			}
		}
	}()
}

func (c CpuCollector) collectAndSendMetric(t time.Time, metric chan<- types.Metric) error {
	cpuPercent, err := cpu.Percent(0, false)
	if err != nil {
		return err
	}

	select {
	case metric <- types.Metric{
		Type:      "cpu",
		Values:    map[string]any{"cpuPercent": cpuPercent[0]},
		Timestamp: t,
	}:
		return nil
	default:
		return fmt.Errorf("metric channel full")
	}
}
