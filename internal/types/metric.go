package types

import (
	"time"
)

type Metric struct {
	Timestamp time.Time      `json:"timestamp"`
	Type      string         `json:"type"`
	Values    map[string]any `json:"values"`
}
