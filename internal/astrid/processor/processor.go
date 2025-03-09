package processor

import (
	"context"
	"log"
	"sync"

	"github.com/anshulkanwar/night-fury-radar/internal/astrid/collectors"
	"github.com/anshulkanwar/night-fury-radar/internal/storage"
	"github.com/anshulkanwar/night-fury-radar/internal/types"
)

type MetricProcessor struct {
	ctx        context.Context
	storage    *storage.Storage
	metric     chan types.Metric
	collectors []collectors.MetricCollector
	wg         sync.WaitGroup
	cancel     context.CancelFunc
}

func NewMetricProcessor() *MetricProcessor {
	ctx, cancel := context.WithCancel(context.Background())
	return &MetricProcessor{
		ctx:     ctx,
		storage: storage.NewStorage(),
		cancel:  cancel,
		metric:  make(chan types.Metric, 100),
	}
}

func (p *MetricProcessor) AddCollector(collector collectors.MetricCollector) {
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
				p.storage.Store(m)
			}
		}
	}()
}

func (p *MetricProcessor) Stop() {
	p.cancel()
	p.wg.Wait()
	close(p.metric)
	p.storage.Close()
}
