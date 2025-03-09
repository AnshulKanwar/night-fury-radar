package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/anshulkanwar/night-fury-radar/internal/types"
	_ "github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

func NewStorage() *Storage {
	dbUser := os.Getenv("DB_USER")
	dbName := os.Getenv("DB_NAME")
	connStr := fmt.Sprintf("user=%s dbname=%s sslmode=disable", dbUser, dbName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	return &Storage{
		db: db,
	}
}

func (s *Storage) Close() {
	s.db.Close()
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
