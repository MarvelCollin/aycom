package db

import (
	"log"

	"gorm.io/gorm"
)

func UpdatePremiumRequestsTable(db *gorm.DB) error {
	log.Println("Checking premium_requests table structure...")

	var columnExists bool
	err := db.Raw("SELECT EXISTS (SELECT FROM information_schema.columns WHERE table_name = 'premium_requests' AND column_name = 'identity_card_number')").Scan(&columnExists).Error
	if err != nil {
		log.Printf("Error checking if identity_card_number column exists: %v", err)
		return err
	}

	if !columnExists {
		log.Println("Adding missing identity_card_number column to premium_requests table...")
		if err := db.Exec("ALTER TABLE premium_requests ADD COLUMN identity_card_number TEXT").Error; err != nil {
			log.Printf("Error adding identity_card_number column: %v", err)
			return err
		}
		log.Println("Successfully added identity_card_number column to premium_requests table")
	} else {
		log.Println("identity_card_number column already exists")
	}

	err = db.Raw("SELECT EXISTS (SELECT FROM information_schema.columns WHERE table_name = 'premium_requests' AND column_name = 'face_photo_url')").Scan(&columnExists).Error
	if err != nil {
		log.Printf("Error checking if face_photo_url column exists: %v", err)
		return err
	}

	if !columnExists {
		log.Println("Adding missing face_photo_url column to premium_requests table...")
		if err := db.Exec("ALTER TABLE premium_requests ADD COLUMN face_photo_url TEXT").Error; err != nil {
			log.Printf("Error adding face_photo_url column: %v", err)
			return err
		}
		log.Println("Successfully added face_photo_url column to premium_requests table")
	} else {
		log.Println("face_photo_url column already exists")
	}

	log.Println("Premium requests table structure update completed")
	return nil
}