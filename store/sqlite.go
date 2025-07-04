package store

import (
	"encoding/json"
	"time"

	"simple/types"

	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Threads represents a thread record stored in the SQLite database.
type Threads struct {
	ID        string `gorm:"primaryKey"`
	Title     string
	Status    string
	Labels    datatypes.JSON `gorm:"type:json"`
	Customer  string
	Company   string
	CreatedAt *time.Time `gorm:"autoCreateTime:false"`
	UpdatedAt *time.Time `gorm:"autoUpdateTime:false"`
}

// SQLiteInitDB initializes the SQLite database connection using GORM.
// It accepts a file path (for example, "threads.db") and auto-migrates the Threads table.
func SQLiteInitDB(dbPath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto-migrate the schema to keep it up to date.
	if err := db.AutoMigrate(&Threads{}); err != nil {
		return nil, err
	}

	return db, nil
}

// SQLiteSaveThreads saves a slice of Threads objects into the database.
// It iterates over each thread and upserts the record.
func SQLiteSaveThreads(db *gorm.DB, threads []*Threads) error {
	// Iterate over each thread report and save it.
	for _, thread := range threads {
		if err := db.Save(thread).Error; err != nil {
			return err
		}
	}
	return nil
}

// SQLiteFromThread converts a thread from the API (simple/types.Thread) into a Threads model.
// It processes labels, customer, company, and times.
func SQLiteFromThread(t *types.Thread) *Threads {
	// Process labels: marshal multiple labels into a JSON array.
	var labelsJSON datatypes.JSON
	if t.Labels != nil && len(t.Labels) > 0 {
		var labelNames []string
		for _, label := range t.Labels {
			labelNames = append(labelNames, label.LabelType.Name)
		}
		if b, err := json.Marshal(labelNames); err == nil {
			labelsJSON = datatypes.JSON(b)
		} else {
			// fallback to empty JSON array on error
			labelsJSON = datatypes.JSON([]byte("[]"))
		}
	} else {
		labelsJSON = datatypes.JSON([]byte("[]"))
	}

	// Process customer and company.
	customer := ""
	company := ""
	if t.Customer != nil {
		customer = t.Customer.FullName
		if t.Customer.Company != nil {
			company = t.Customer.Company.Name
		}
	}

	// Process created_at timestamp.
	var created *time.Time
	if t.CreatedAt != nil {
		if tm, err := t.CreatedAt.Time(); err == nil {
			created = &tm
		}
	}

	// Process updated_at timestamp.
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
