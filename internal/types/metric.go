package types

import (
	"time"
)

type Metric struct {
	Timestamp time.Time
	Type      string
	Values    map[string]any
}
