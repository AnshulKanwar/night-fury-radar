package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/anshulkanwar/night-fury-radar/internal/types"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type Storage struct {
	db       *sql.DB
	Listener *pq.Listener
}

func NewStorage() *Storage {
	dbUser := os.Getenv("DB_USER")
	dbName := os.Getenv("DB_NAME")
	connStr := fmt.Sprintf("user=%s dbname=%s sslmode=disable", dbUser, dbName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	reportProblem := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			log.Fatal(err)
		}
	}

	listener := pq.NewListener(connStr, 10*time.Second, time.Minute, reportProblem)
	err = listener.Listen("metrics")
	if err != nil {
		log.Fatal(err)
	}

	return &Storage{
		db:       db,
		Listener: listener,
	}
}

func (s *Storage) Close() {
	s.db.Close()
	s.Listener.Close()
}

func (s *Storage) Store(metric types.Metric) {
	jsonValues, err := json.Marshal(metric.Values)
	if err != nil {
		log.Fatal(err)
	}

	_, err = s.db.Exec("INSERT INTO system_metrics (timestamp, type, values) VALUES ($1, $2, $3)", metric.Timestamp, metric.Type, jsonValues)
	if err != nil {
		log.Fatal(err)
	}
}

func (s *Storage) ReadLast100Points(metricType string) []types.Metric {
	rows, err := s.db.Query("SELECT type, timestamp, values FROM system_metrics WHERE type = $1 ORDER BY timestamp DESC LIMIT 100", metricType)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	metrics := make([]types.Metric, 0)
	for rows.Next() {
		var metric types.Metric
		var valuesJSON []byte
		if err := rows.Scan(&metric.Type, &metric.Timestamp, &valuesJSON); err != nil {
			log.Fatal(err)
		}
		if err := json.Unmarshal(valuesJSON, &metric.Values); err != nil {
			log.Fatal(err)
		}
		metrics = append(metrics, metric)
	}

	return metrics
}
