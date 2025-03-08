package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
)

type Metric struct {
	Name      string
	Tags      map[string]string
	Fields    map[string]any
	Timestamp time.Time
}

type MetricCollector interface {
	Start(ctx context.Context, metric chan Metric, wg *sync.WaitGroup)
}

type CpuCollector struct{}

func (c CpuCollector) Start(ctx context.Context, metric chan Metric, wg *sync.WaitGroup) {
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

func (c CpuCollector) collectAndSendMetric(t time.Time, metric chan<- Metric) error {
	cpuPercent, err := cpu.Percent(0, false)
	if err != nil {
		return err
	}

	select {
	case metric <- Metric{
		Name:      "system_metric",
		Tags:      map[string]string{"type": "cpu"},
		Fields:    map[string]any{"cpu_percent": cpuPercent[0]},
		Timestamp: t,
	}:
		return nil
	default:
		return fmt.Errorf("metric channel full")
	}
}

type MetricProcessor struct {
	ctx        context.Context
	metric     chan Metric
	collectors []MetricCollector
	wg         sync.WaitGroup
	cancel     context.CancelFunc
}

func NewMetricProcessor() *MetricProcessor {
	ctx, cancel := context.WithCancel(context.Background())
	return &MetricProcessor{
		ctx:    ctx,
		cancel: cancel,
		metric: make(chan Metric, 100),
	}
}

func (p *MetricProcessor) AddCollector(collector MetricCollector) {
	p.collectors = append(p.collectors, collector)
}

func (p *MetricProcessor) Start() {
	for _, collector := range p.collectors {
		collector.Start(p.ctx, p.metric, &p.wg)
	}

	p.wg.Add(1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recovered from panic: %v", r)
			}
			p.wg.Done()
		}()
		for {
			select {
			case <-p.ctx.Done():
				return
			case m := <-p.metric:
				log.Println(m)
			}
		}
	}()
}

func (p *MetricProcessor) Stop() {
	p.cancel()
	p.wg.Wait()
	close(p.metric)
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	processor := NewMetricProcessor()
	processor.AddCollector(CpuCollector{})

	processor.Start()

	<-ctx.Done()
	log.Println("Shutting down gracefully...")
	processor.Stop()
}
