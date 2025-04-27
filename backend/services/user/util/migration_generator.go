package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

// MigrationTemplate is the template for Go-based migrations
const MigrationTemplate = `package migrations

import (
	"database/sql"

	"github.com/pressly/goose/v3"
	"gorm.io/gorm"

	"github.com/Acad600-Tpa/WEB-MV-242/services/user/model"
)

func init() {
	goose.AddMigration(Up{{.Timestamp}}, Down{{.Timestamp}})
}

// Up{{.Timestamp}} creates the {{.TableName}} table
func Up{{.Timestamp}}(tx *sql.Tx) error {
	db, err := convertTxToGorm(tx)
	if err != nil {
		return err
	}

	// Create table based on GORM model
	if err := db.AutoMigrate(&model.{{.ModelName}}{}); err != nil {
		return err
	}

	// Create indexes for performance
{{range .Indexes}}	if err := db.Exec("CREATE INDEX IF NOT EXISTS {{.}} ON {{$.TableName}}({{.Field}})").Error; err != nil {
		return err
	}
{{end}}
	return nil
}

// Down{{.Timestamp}} drops the {{.TableName}} table
func Down{{.Timestamp}}(tx *sql.Tx) error {
	db, err := convertTxToGorm(tx)
	if err != nil {
		return err
	}

	return db.Migrator().DropTable("{{.TableName}}")
}
`

// MigrationData holds data for the migration template
type MigrationData struct {
	Timestamp string
	ModelName string
	TableName string
	Indexes   []IndexData
}

// IndexData holds data for index creation
type IndexData struct {
	Name  string
	Field string
}

// GenerateMigration creates a new migration file for a model
func GenerateMigration(modelName, tableName string, indexes []IndexData) error {
	// Create timestamp in the format YYYYMMDDHHMMSS
	timestamp := time.Now().Format("20060102150405")

	// Prepare data for template
	data := MigrationData{
		Timestamp: timestamp,
		ModelName: modelName,
		TableName: tableName,
		Indexes:   indexes,
	}

	// Parse template
	tmpl, err := template.New("migration").Parse(MigrationTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Create migrations directory if it doesn't exist
	migrationsDir := filepath.Join("migrations")
	if err := os.MkdirAll(migrationsDir, 0755); err != nil {
		return fmt.Errorf("failed to create migrations directory: %w", err)
	}

	// Create migration file
	filename := filepath.Join(migrationsDir, fmt.Sprintf("%s_create_%s.go", timestamp, strings.ToLower(tableName)))
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create migration file: %w", err)
	}
	defer file.Close()

	// Execute template
	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to write migration file: %w", err)
	}

	fmt.Printf("Migration file created: %s\n", filename)
	return nil
}
