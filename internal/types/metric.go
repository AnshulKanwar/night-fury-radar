package types

import (
	"time"
)

type Metric struct {
	Name      string
	Tags      map[string]string
	Fields    map[string]any
	Timestamp time.Time
}
