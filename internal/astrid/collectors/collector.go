package collectors

import (
	"context"
	"sync"

	"github.com/anshulkanwar/night-fury-radar/internal/types"
)

type MetricCollector interface {
	Start(ctx context.Context, metric chan types.Metric, wg *sync.WaitGroup)
}
