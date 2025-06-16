package db

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CommunitySeeder struct {
	db *gorm.DB
}

func NewCommunitySeeder(db *gorm.DB) *CommunitySeeder {
	return &CommunitySeeder{
		db: db,
	}
}

func (s *CommunitySeeder) SeedAll() error {
	log.Println("Starting community seeder...")

	// Get current user IDs from the system
	var userIDs []uuid.UUID
	if err := s.fetchActiveUserIDs(&userIDs); err != nil {
		log.Printf("Warning: Could not fetch user IDs: %v. Using default IDs.", err)
		// Use some default IDs if we can't fetch real ones
		userIDs = []uuid.UUID{
			uuid.MustParse("91df5727-a9c5-427e-94ce-e0486e3bfdb7"), // Current user ID from logs
			uuid.MustParse("fd434c0e-95de-41d0-a576-9d4ea2fed7e9"), // Another ID seen in logs
		}
	}

	if len(userIDs) == 0 {
		log.Printf("Warning: No user IDs available. Using default IDs.")
		userIDs = []uuid.UUID{
			uuid.MustParse("91df5727-a9c5-427e-94ce-e0486e3bfdb7"),
			uuid.MustParse("fd434c0e-95de-41d0-a576-9d4ea2fed7e9"),
		}
	}

	log.Printf("Using user IDs for seeding: %v", userIDs)

	if err := s.SeedCommunities(userIDs); err != nil {
		return err
	}
	if err := s.SeedCommunityMembers(userIDs); err != nil {
		return err
	}
	if err := s.SeedJoinRequests(userIDs); err != nil {
		return err
	}

	log.Println("Community seeding completed successfully!")
	return nil
}

// Fetch active user IDs from the user service database via a postgres connection
func (s *CommunitySeeder) fetchActiveUserIDs(userIDs *[]uuid.UUID) error {
	// Create a temporary struct to hold user IDs
	type UserID struct {
		ID uuid.UUID
	}
	var users []UserID

	// Try to get user IDs from our own DB first (from existing members or pending requests)
	if err := s.db.Raw("SELECT DISTINCT user_id as id FROM community_members WHERE deleted_at IS NULL LIMIT 10").Scan(&users).Error; err != nil {
		log.Printf("Failed to fetch user IDs from community members: %v", err)
	}

	if len(users) == 0 {
		if err := s.db.Raw("SELECT DISTINCT user_id as id FROM community_join_requests WHERE deleted_at IS NULL LIMIT 10").Scan(&users).Error; err != nil {
			log.Printf("Failed to fetch user IDs from join requests: %v", err)
		}
	}

	// If we found some, use them
	if len(users) > 0 {
		for _, user := range users {
			*userIDs = append(*userIDs, user.ID)
		}
		return nil
	}

	// Otherwise, return an error
	return fmt.Errorf("no user IDs found in database")
}

func (s *CommunitySeeder) SeedCommunities(userIDs []uuid.UUID) error {
	var count int64
	s.db.Table("communities").Count(&count)
	if count > 0 {
		log.Println("Communities already exist, skipping seeding")
		return nil
	}

	// Use the first user ID for creating communities
	creatorID := userIDs[0]
	log.Printf("Using creator ID for communities: %s", creatorID)

	type Community struct {
		CommunityID uuid.UUID  `gorm:"type:uuid;primaryKey;column:community_id"`
		Name        string     `gorm:"type:varchar(100);unique;not null"`
		Description string     `gorm:"type:text;not null"`
		LogoURL     string     `gorm:"type:varchar(512);not null"`
		BannerURL   string     `gorm:"type:varchar(512);not null"`
		CreatorID   uuid.UUID  `gorm:"type:uuid;not null"`
		IsApproved  bool       `gorm:"default:false;not null"`
		CreatedAt   time.Time  `gorm:"autoCreateTime"`
		UpdatedAt   time.Time  `gorm:"autoUpdateTime"`
		DeletedAt   *time.Time `gorm:"index"`
	}

	communities := []Community{
		{
			CommunityID: uuid.New(),
			Name:        "Tech Enthusiasts",
			Description: "A community for technology lovers and early adopters. We discuss the latest gadgets, software releases, and tech trends.",
			LogoURL:     "https://via.placeholder.com/150",
			BannerURL:   "https://via.placeholder.com/600x200",
			CreatorID:   creatorID,
			IsApproved:  true,
			CreatedAt:   time.Now().Add(-60 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-60 * 24 * time.Hour),
		},
		{
			CommunityID: uuid.New(),
			Name:        "Fitness & Health",
			Description: "Join us to discuss fitness routines, health tips, nutrition advice, and wellness strategies.",
			LogoURL:     "https://via.placeholder.com/150",
			BannerURL:   "https://via.placeholder.com/600x200",
			CreatorID:   creatorID,
			IsApproved:  true,
			CreatedAt:   time.Now().Add(-55 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-55 * 24 * time.Hour),
		},
		{
			CommunityID: uuid.New(),
			Name:        "Developers Hub",
			Description: "A community for software developers to share knowledge, discuss programming languages, and collaborate on projects.",
			LogoURL:     "https://via.placeholder.com/150",
			BannerURL:   "https://via.placeholder.com/600x200",
			CreatorID:   creatorID,
			IsApproved:  true,
			CreatedAt:   time.Now().Add(-50 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-50 * 24 * time.Hour),
		},
	}

	if err := s.db.Table("communities").Create(&communities).Error; err != nil {
		return fmt.Errorf("failed to create communities: %w", err)
	}

	log.Printf("Created %d communities", len(communities))
	return nil
}

func (s *CommunitySeeder) SeedCommunityMembers(userIDs []uuid.UUID) error {
	// First check if there are already community members
	var count int64
	s.db.Table("community_members").Count(&count)
	if count > 0 {
		log.Println("Community members already exist, truncating and reseeding")
		// Remove existing records to start fresh
		if err := s.db.Exec("DELETE FROM community_members").Error; err != nil {
			log.Printf("Warning: Failed to delete existing members: %v", err)
		}
	}

	var communities []struct {
		CommunityID uuid.UUID
		Name        string
		CreatorID   uuid.UUID
	}

	if err := s.db.Table("communities").Select("community_id, name, creator_id").Find(&communities).Error; err != nil {
		return fmt.Errorf("failed to fetch communities: %w", err)
	}

	if len(communities) == 0 {
		return fmt.Errorf("no communities found to seed members for")
	}

	log.Printf("Found %d communities to create members for", len(communities))

	type CommunityMember struct {
		CommunityID uuid.UUID  `gorm:"type:uuid;primaryKey;column:community_id"`
		UserID      uuid.UUID  `gorm:"type:uuid;primaryKey;column:user_id"`
		Role        string     `gorm:"type:varchar(10);not null"`
		CreatedAt   time.Time  `gorm:"autoCreateTime"`
		UpdatedAt   time.Time  `gorm:"autoUpdateTime"`
		DeletedAt   *time.Time `gorm:"index"`
	}

	members := []CommunityMember{}

	// First, add the creator as admin for each community
	for _, community := range communities {
		members = append(members, CommunityMember{
			CommunityID: community.CommunityID,
			UserID:      community.CreatorID,
			Role:        "admin",
			CreatedAt:   time.Now().Add(-60 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-60 * 24 * time.Hour),
		})

		log.Printf("Added admin member for community %s: %s", community.Name, community.CreatorID)
	}

	// Then add other users as members to some communities
	for _, userID := range userIDs {
		// Skip if this is the creator (already added as admin)
		isCreator := false
		for _, community := range communities {
			if community.CreatorID == userID {
				isCreator = true
				break
			}
		}

		if isCreator {
			continue
		}

		// Add this user as a member to the first community
		if len(communities) > 0 {
			members = append(members, CommunityMember{
				CommunityID: communities[0].CommunityID,
				UserID:      userID,
				Role:        "member",
				CreatedAt:   time.Now().Add(-55 * 24 * time.Hour),
				UpdatedAt:   time.Now().Add(-55 * 24 * time.Hour),
			})

			log.Printf("Added regular member for community %s: %s", communities[0].Name, userID)
		}

		// If there's more than one community, add them as a moderator to the second
		if len(communities) > 1 {
			members = append(members, CommunityMember{
				CommunityID: communities[1].CommunityID,
				UserID:      userID,
				Role:        "moderator",
				CreatedAt:   time.Now().Add(-50 * 24 * time.Hour),
				UpdatedAt:   time.Now().Add(-50 * 24 * time.Hour),
			})

			log.Printf("Added moderator member for community %s: %s", communities[1].Name, userID)
		}
	}

	if len(members) == 0 {
		log.Println("Warning: No community members to seed")
		return nil
	}

	if err := s.db.Table("community_members").Create(&members).Error; err != nil {
		return fmt.Errorf("failed to create community members: %w", err)
	}

	log.Printf("Created %d community members", len(members))
	return nil
}

func (s *CommunitySeeder) SeedJoinRequests(userIDs []uuid.UUID) error {
	// First check if there are already join requests
	var count int64
	s.db.Table("community_join_requests").Count(&count)
	if count > 0 {
		log.Println("Community join requests already exist, truncating and reseeding")
		// Remove existing records to start fresh
		if err := s.db.Exec("DELETE FROM community_join_requests").Error; err != nil {
			log.Printf("Warning: Failed to delete existing join requests: %v", err)
		}
	}

	// If we don't have enough users, return
	if len(userIDs) < 2 {
		log.Println("Not enough users to create join requests")
		return nil
	}

	// Get communities that the second user isn't already a member of
	var availableCommunities []struct {
		CommunityID uuid.UUID
		Name        string
	}

	secondUserID := userIDs[1]

	query := `
		SELECT c.community_id, c.name 
		FROM communities c
		WHERE c.community_id NOT IN (
			SELECT cm.community_id 
			FROM community_members cm 
			WHERE cm.user_id = ? AND cm.deleted_at IS NULL
		)
		LIMIT 2
	`

	if err := s.db.Raw(query, secondUserID).Scan(&availableCommunities).Error; err != nil {
		return fmt.Errorf("failed to fetch available communities: %w", err)
	}

	if len(availableCommunities) == 0 {
		log.Println("No available communities found for join requests")
		return nil
	}

	type CommunityJoinRequest struct {
		RequestID   uuid.UUID  `gorm:"type:uuid;primaryKey;column:request_id"`
		CommunityID uuid.UUID  `gorm:"type:uuid;not null"`
		UserID      uuid.UUID  `gorm:"type:uuid;not null"`
		Status      string     `gorm:"type:varchar(10);not null"`
		CreatedAt   time.Time  `gorm:"autoCreateTime"`
		UpdatedAt   time.Time  `gorm:"autoUpdateTime"`
		DeletedAt   *time.Time `gorm:"index"`
	}

	requests := []CommunityJoinRequest{}

	// Create a pending request for the second user
	for _, community := range availableCommunities {
		requests = append(requests, CommunityJoinRequest{
			RequestID:   uuid.New(),
			CommunityID: community.CommunityID,
			UserID:      secondUserID,
			Status:      "pending",
			CreatedAt:   time.Now().Add(-10 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-10 * 24 * time.Hour),
		})

		log.Printf("Added pending join request for community %s by user %s", community.Name, secondUserID)
	}

	if len(requests) == 0 {
		log.Println("No join requests to seed")
		return nil
	}

	if err := s.db.Table("community_join_requests").Create(&requests).Error; err != nil {
		return fmt.Errorf("failed to create join requests: %w", err)
	}

	log.Printf("Created %d community join requests", len(requests))
	return nil
}
