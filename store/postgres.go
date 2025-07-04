package store

import (
	"encoding/json"
	"fmt"
	"time"

	"simple/types"

	"gorm.io/datatypes"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// PostgresConfig contains the configuration values for connecting to a Postgres server.
// Using context7, the configuration setup follows the same principle as with other components.
type PostgresConfig struct {
	Host     string // Postgres host address
	Port     int    // Postgres port
	User     string // Username for the database
	Password string // Password for the database
	DBName   string // Database name
	SSLMode  string // SSL mode (disable, require, verify-full, etc.)
}

// PostgresInitDB initializes the Postgres database connection using GORM.
// It builds a DSN string from the provided configuration, opens the connection, and auto-migrates the Threads model.
func PostgresInitDB(cfg PostgresConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto-migrate the schema (using the Threads model defined in sqlite.go) to keep it up to date.
	if err := db.AutoMigrate(&Threads{}); err != nil {
		return nil, err
	}

	return db, nil
}

// PostgresSaveThreads saves a slice of Threads objects into the Postgres database.
// It iterates over each thread and upserts the record.
func PostgresSaveThreads(db *gorm.DB, threads []*Threads) error {
	for _, thread := range threads {
		if err := db.Save(thread).Error; err != nil {
			return err
		}
	}
	return nil
}

// PostgresFromThread converts a thread from the API (simple/types.Thread) into a Threads model.
// It processes the labels, customer/company, and timestamps in the same way as SQLiteFromThread, ensuring consistency.
func PostgresFromThread(t *types.Thread) *Threads {
	var labelsJSON datatypes.JSON
	if t.Labels != nil && len(t.Labels) > 0 {
		labelsMap := make(map[string]string)
		for i, label := range t.Labels {
			key := fmt.Sprintf("Label%d", i)
			labelsMap[key] = label.LabelType.Name
		}
		if b, err := json.Marshal(labelsMap); err == nil {
			labelsJSON = datatypes.JSON(b)
		} else {
			labelsJSON = datatypes.JSON([]byte("{}"))
		}
	} else {
		labelsJSON = datatypes.JSON([]byte("{}"))
	}

	customer := ""
	company := ""
	if t.Customer != nil {
		customer = t.Customer.FullName
		if t.Customer.Company != nil {
			company = t.Customer.Company.Name
		}
	}

	var created *time.Time
	if t.CreatedAt != nil {
		if tm, err := t.CreatedAt.Time(); err == nil {
			created = &tm
		}
	}

	var updated *time.Time
	if t.UpdatedAt != nil {
		if tm, err := t.UpdatedAt.Time(); err == nil {
			updated = &tm
		}
	}

	return &Threads{
		ID:        t.ID,
		Title:     t.Title,
		Status:    t.Status,
		Labels:    labelsJSON,
		Customer:  customer,
		Company:   company,
		CreatedAt: created,
		UpdatedAt: updated,
	}
}
