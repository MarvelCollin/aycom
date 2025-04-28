package seeds

import (
	"log"

	"github.com/Acad600-Tpa/WEB-MV-242/backend/services/user/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SeedAll runs all seeders
func SeedAll(db *gorm.DB) error {
	if err := SeedUserProfiles(db); err != nil {
		return err
	}

	if err := SeedContacts(db); err != nil {
		return err
	}

	return nil
}

// SeedUserProfiles adds test user profiles to the database
func SeedUserProfiles(db *gorm.DB) error {
	// Check if profiles already exist
	var count int64
	db.Model(&model.UserProfile{}).Count(&count)
	if count > 0 {
		log.Println("UserProfiles table already has data, skipping seeding")
		return nil
	}

	// Fixed UUIDs matching the auth service users
	adminID := uuid.MustParse("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11")
	testUserID := uuid.MustParse("b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12")

	profileID1 := uuid.MustParse("d0eebc99-9c0b-4ef8-bb6d-6bb9bd380a14")
	profileID2 := uuid.MustParse("e0eebc99-9c0b-4ef8-bb6d-6bb9bd380a15")

	// Create test user profiles
	profiles := []model.UserProfile{
		{
			ID:                      profileID1,
			UserID:                  adminID,
			Bio:                     "Admin user with system privileges",
			ProfilePictureURL:       "https://example.com/avatars/admin.jpg",
			BannerURL:               "https://example.com/banners/admin_banner.jpg",
			Location:                "Jakarta, Indonesia",
			Website:                 "https://admin.example.com",
			SocialLinks:             model.JSONB{"twitter": "@admin", "github": "admin-github"},
			Interests:               model.StringArray{"coding", "system administration", "security"},
			Language:                "en",
			Theme:                   "dark",
			IsPrivate:               false,
			IsPremium:               true,
			NotificationPreferences: model.JSONB{"email": true, "push": true},
		},
		{
			ID:                      profileID2,
			UserID:                  testUserID,
			Bio:                     "Regular test user account",
			ProfilePictureURL:       "https://example.com/avatars/testuser.jpg",
			BannerURL:               "https://example.com/banners/testuser_banner.jpg",
			Location:                "Bandung, Indonesia",
			Website:                 "https://testuser.example.com",
			SocialLinks:             model.JSONB{"instagram": "@testuser", "linkedin": "test-user"},
			Interests:               model.StringArray{"reading", "travel", "photography"},
			Language:                "id",
			Theme:                   "light",
			IsPrivate:               true,
			IsPremium:               false,
			NotificationPreferences: model.JSONB{"email": true, "push": false},
		},
	}

	result := db.Create(&profiles)
	if result.Error != nil {
		return result.Error
	}

	log.Printf("Created %d test user profiles", len(profiles))
	return nil
}

// SeedContacts adds test contacts between users
func SeedContacts(db *gorm.DB) error {
	// Check if contacts already exist
	var count int64
	db.Model(&model.Contact{}).Count(&count)
	if count > 0 {
		log.Println("Contacts table already has data, skipping seeding")
		return nil
	}

	// Fixed UUIDs matching the auth service users
	adminID := uuid.MustParse("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11")
	testUserID := uuid.MustParse("b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12")

	contactID1 := uuid.MustParse("f0eebc99-9c0b-4ef8-bb6d-6bb9bd380a16")
	contactID2 := uuid.MustParse("g0eebc99-9c0b-4ef8-bb6d-6bb9bd380a17")

	// Create test contacts
	contacts := []model.Contact{
		{
			ID:            contactID1,
			UserID:        adminID,
			ContactUserID: testUserID,
			Relationship:  "friend",
		},
		{
			ID:            contactID2,
			UserID:        testUserID,
			ContactUserID: adminID,
			Relationship:  "friend",
		},
	}

	result := db.Create(&contacts)
	if result.Error != nil {
		return result.Error
	}

	log.Printf("Created %d test contacts", len(contacts))
	return nil
}
