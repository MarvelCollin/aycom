package db

import (
	"log"
	"time"

	"aycom/backend/services/thread/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CheckUserDataExists verifies that user data exists before seeding threads
func CheckUserDataExists(db *gorm.DB) bool {
	adminID := uuid.MustParse("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11")

	// Try to find the admin user using GORM
	var count int64
	db.Table("users").Where("id = ?", adminID).Count(&count)

	if count == 0 {
		log.Println("Warning: User data not found. Make sure to run user migrations and seeders first.")
		return false
	}

	return true
}

// SeedDatabase seeds the database with initial data
func SeedDatabase(db *gorm.DB) error {
	log.Println("Starting thread database seeding...")

	// Check if threads already exist
	var count int64
	db.Model(&model.Thread{}).Count(&count)
	if count > 0 {
		log.Println("Threads already exist, skipping seeding")
		return nil
	}

	// Check if user data exists
	if !CheckUserDataExists(db) {
		log.Println("Unable to seed thread data: user data doesn't exist yet")
		return nil
	}

	// Predefined user IDs from user_seeder.go
	adminID := uuid.MustParse("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11")
	johnID := uuid.MustParse("b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12")
	janeID := uuid.MustParse("c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a13")
	samID := uuid.MustParse("d0eebc99-9c0b-4ef8-bb6d-6bb9bd380a14")

	// Create sample threads
	threads := []model.Thread{
		// Admin threads
		{
			ThreadID:  uuid.New(),
			UserID:    adminID,
			Content:   "Welcome to AYCOM! This is our first official announcement.",
			IsPinned:  true,
			CreatedAt: time.Now().Add(-48 * time.Hour),
			UpdatedAt: time.Now().Add(-48 * time.Hour),
		},
		{
			ThreadID:  uuid.New(),
			UserID:    adminID,
			Content:   "We're excited to announce new features coming soon!",
			CreatedAt: time.Now().Add(-24 * time.Hour),
			UpdatedAt: time.Now().Add(-24 * time.Hour),
		},

		// John's threads
		{
			ThreadID:  uuid.New(),
			UserID:    johnID,
			Content:   "Hello everyone! Just joined this platform and it looks amazing.",
			CreatedAt: time.Now().Add(-36 * time.Hour),
			UpdatedAt: time.Now().Add(-36 * time.Hour),
		},
		{
			ThreadID:  uuid.New(),
			UserID:    johnID,
			Content:   "What's everyone working on today? #coding #programming",
			CreatedAt: time.Now().Add(-12 * time.Hour),
			UpdatedAt: time.Now().Add(-12 * time.Hour),
		},

		// Jane's threads
		{
			ThreadID:  uuid.New(),
			UserID:    janeID,
			Content:   "Just finished my new project. Check it out!",
			CreatedAt: time.Now().Add(-18 * time.Hour),
			UpdatedAt: time.Now().Add(-18 * time.Hour),
		},
		{
			ThreadID:  uuid.New(),
			UserID:    janeID,
			Content:   "Looking for collaborators on a new open source project. DM me if interested.",
			CreatedAt: time.Now().Add(-6 * time.Hour),
			UpdatedAt: time.Now().Add(-6 * time.Hour),
		},

		// Sam's threads
		{
			ThreadID:  uuid.New(),
			UserID:    samID,
			Content:   "First day at my new job! Excited for this new chapter.",
			CreatedAt: time.Now().Add(-30 * time.Hour),
			UpdatedAt: time.Now().Add(-30 * time.Hour),
		},
		{
			ThreadID:  uuid.New(),
			UserID:    samID,
			Content:   "Anyone else attending the tech conference next week? #techconf",
			CreatedAt: time.Now().Add(-8 * time.Hour),
			UpdatedAt: time.Now().Add(-8 * time.Hour),
		},
	}

	// Insert threads
	result := db.Create(&threads)
	if result.Error != nil {
		log.Printf("Error seeding threads: %v", result.Error)
		return result.Error
	}

	// Create some hashtags
	hashtags := []model.Hashtag{
		{
			HashtagID: uuid.New(),
			Text:      "coding",
			CreatedAt: time.Now(),
		},
		{
			HashtagID: uuid.New(),
			Text:      "programming",
			CreatedAt: time.Now(),
		},
		{
			HashtagID: uuid.New(),
			Text:      "opensource",
			CreatedAt: time.Now(),
		},
		{
			HashtagID: uuid.New(),
			Text:      "techconf",
			CreatedAt: time.Now(),
		},
	}

	// Insert hashtags
	result = db.Create(&hashtags)
	if result.Error != nil {
		log.Printf("Error seeding hashtags: %v", result.Error)
		return result.Error
	}

	// Create thread-hashtag associations
	threadHashtags := []model.ThreadHashtag{
		{
			ThreadID:  threads[3].ThreadID,   // John's second thread
			HashtagID: hashtags[0].HashtagID, // coding
		},
		{
			ThreadID:  threads[3].ThreadID,   // John's second thread
			HashtagID: hashtags[1].HashtagID, // programming
		},
		{
			ThreadID:  threads[5].ThreadID,   // Jane's second thread
			HashtagID: hashtags[2].HashtagID, // opensource
		},
		{
			ThreadID:  threads[7].ThreadID,   // Sam's second thread
			HashtagID: hashtags[3].HashtagID, // techconf
		},
	}

	// Insert thread-hashtag associations
	result = db.Create(&threadHashtags)
	if result.Error != nil {
		log.Printf("Error seeding thread hashtags: %v", result.Error)
		return result.Error
	}

	// Create some replies
	replies := []model.Reply{
		{
			ReplyID:   uuid.New(),
			ThreadID:  threads[0].ThreadID, // Admin's first thread
			UserID:    johnID,
			Content:   "Excited to be here!",
			CreatedAt: time.Now().Add(-47 * time.Hour),
			UpdatedAt: time.Now().Add(-47 * time.Hour),
		},
		{
			ReplyID:   uuid.New(),
			ThreadID:  threads[2].ThreadID, // John's first thread
			UserID:    janeID,
			Content:   "Welcome John! Nice to meet you.",
			CreatedAt: time.Now().Add(-35 * time.Hour),
			UpdatedAt: time.Now().Add(-35 * time.Hour),
		},
		{
			ReplyID:   uuid.New(),
			ThreadID:  threads[4].ThreadID, // Jane's first thread
			UserID:    johnID,
			Content:   "That looks amazing! Great work.",
			CreatedAt: time.Now().Add(-17 * time.Hour),
			UpdatedAt: time.Now().Add(-17 * time.Hour),
		},
		{
			ReplyID:   uuid.New(),
			ThreadID:  threads[6].ThreadID, // Sam's first thread
			UserID:    adminID,
			Content:   "Congratulations on the new job!",
			CreatedAt: time.Now().Add(-29 * time.Hour),
			UpdatedAt: time.Now().Add(-29 * time.Hour),
		},
	}

	// Insert replies
	result = db.Create(&replies)
	if result.Error != nil {
		log.Printf("Error seeding replies: %v", result.Error)
		return result.Error
	}

	// Create some likes
	likes := []model.Like{
		{
			UserID:    adminID,
			ThreadID:  &threads[2].ThreadID, // John's first thread
			CreatedAt: time.Now().Add(-34 * time.Hour),
		},
		{
			UserID:    johnID,
			ThreadID:  &threads[4].ThreadID, // Jane's first thread
			CreatedAt: time.Now().Add(-17 * time.Hour),
		},
		{
			UserID:    janeID,
			ThreadID:  &threads[3].ThreadID, // John's second thread
			CreatedAt: time.Now().Add(-11 * time.Hour),
		},
		{
			UserID:    samID,
			ThreadID:  &threads[5].ThreadID, // Jane's second thread
			CreatedAt: time.Now().Add(-5 * time.Hour),
		},
		{
			UserID:    janeID,
			ThreadID:  &threads[6].ThreadID, // Sam's first thread
			CreatedAt: time.Now().Add(-28 * time.Hour),
		},
	}

	// Insert likes
	result = db.Create(&likes)
	if result.Error != nil {
		log.Printf("Error seeding likes: %v", result.Error)
		return result.Error
	}

	log.Printf("Successfully seeded %d threads, %d hashtags, %d replies, and %d likes",
		len(threads), len(hashtags), len(replies), len(likes))
	return nil
}
