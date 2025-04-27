package migrations

import (
	"database/sql"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// convertTxToGorm converts a sql.Tx to a *gorm.DB instance
func convertTxToGorm(tx *sql.Tx) (*gorm.DB, error) {
	// Create a new GORM DB instance that uses the transaction
	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn: tx,
	}), &gorm.Config{})

	if err != nil {
		return nil, err
	}

	return db, nil
}
