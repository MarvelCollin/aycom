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
	if err := s.SeedCommunities(); err != nil {
		return err
	}
	if err := s.SeedCommunityMembers(); err != nil {
		return err
	}
	if err := s.SeedJoinRequests(); err != nil {
		return err
	}
	return nil
}

func (s *CommunitySeeder) SeedCommunities() error {
	var count int64
	s.db.Table("communities").Count(&count)
	if count > 0 {
		log.Println("Communities already exist, skipping seeding")
		return nil
	}

	adminID := uuid.MustParse("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11")
	johnID := uuid.MustParse("b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12")
	janeID := uuid.MustParse("c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a13")
	samID := uuid.MustParse("d0eebc99-9c0b-4ef8-bb6d-6bb9bd380a14")

	var additionalUsers []struct {
		ID       uuid.UUID
		Username string
	}
	s.db.Table("users").Select("id, username").Where("username IN ?", []string{"techguru", "fitnesscoach"}).Find(&additionalUsers)

	userMap := make(map[string]uuid.UUID)
	for _, user := range additionalUsers {
		userMap[user.Username] = user.ID
	}

	type Community struct {
		CommunityID uuid.UUID  `gorm:"type:uuid;primaryKey;column:community_id"`
		Name        string     `gorm:"type:varchar(100);unique;not null"`
		Description string     `gorm:"type:text;not null"`
		LogoURL     string     `gorm:"type:varchar(512)"`
		BannerURL   string     `gorm:"type:varchar(512)"`
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
			LogoURL:     "https://example.com/logos/tech.png",
			BannerURL:   "https://example.com/banners/tech.png",
			CreatorID:   userMap["techguru"], 
			IsApproved:  true,
			CreatedAt:   time.Now().Add(-60 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-60 * 24 * time.Hour),
		},
		{
			CommunityID: uuid.New(),
			Name:        "Fitness & Health",
			Description: "Join us to discuss fitness routines, health tips, nutrition advice, and wellness strategies.",
			LogoURL:     "https://example.com/logos/fitness.png",
			BannerURL:   "https://example.com/banners/fitness.png",
			CreatorID:   userMap["fitnesscoach"], 
			IsApproved:  true,
			CreatedAt:   time.Now().Add(-55 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-55 * 24 * time.Hour),
		},
		{
			CommunityID: uuid.New(),
			Name:        "Developers Hub",
			Description: "A community for software developers to share knowledge, discuss programming languages, and collaborate on projects.",
			LogoURL:     "https://example.com/logos/dev.png",
			BannerURL:   "https://example.com/banners/dev.png",
			CreatorID:   johnID,
			IsApproved:  true,
			CreatedAt:   time.Now().Add(-50 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-50 * 24 * time.Hour),
		},
		{
			CommunityID: uuid.New(),
			Name:        "Design Showcase",
			Description: "Share your design work, get feedback, and find inspiration from other designers around the world.",
			LogoURL:     "https://example.com/logos/design.png",
			BannerURL:   "https://example.com/banners/design.png",
			CreatorID:   janeID,
			IsApproved:  true,
			CreatedAt:   time.Now().Add(-45 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-45 * 24 * time.Hour),
		},
		{
			CommunityID: uuid.New(),
			Name:        "Book Club",
			Description: "Discuss your favorite books, share recommendations, and join monthly reading challenges.",
			LogoURL:     "https://example.com/logos/books.png",
			BannerURL:   "https://example.com/banners/books.png",
			CreatorID:   samID,
			IsApproved:  true,
			CreatedAt:   time.Now().Add(-40 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-40 * 24 * time.Hour),
		},
		{
			CommunityID: uuid.New(),
			Name:        "Travel Explorers",
			Description: "Share travel tips, photos, and experiences from around the world. Connect with fellow travelers!",
			LogoURL:     "https://example.com/logos/travel.png",
			BannerURL:   "https://example.com/banners/travel.png",
			CreatorID:   adminID,
			IsApproved:  true,
			CreatedAt:   time.Now().Add(-35 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-35 * 24 * time.Hour),
		},
		{
			CommunityID: uuid.New(),
			Name:        "Foodie Paradise",
			Description: "For food lovers to share recipes, restaurant reviews, and cooking tips.",
			LogoURL:     "https://example.com/logos/food.png",
			BannerURL:   "https://example.com/banners/food.png",
			CreatorID:   johnID,
			IsApproved:  true,
			CreatedAt:   time.Now().Add(-30 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-30 * 24 * time.Hour),
		},
		{
			CommunityID: uuid.New(),
			Name:        "Photography Club",
			Description: "Share your photography, learn new techniques, and discuss the latest camera gear.",
			LogoURL:     "https://example.com/logos/photo.png",
			BannerURL:   "https://example.com/banners/photo.png",
			CreatorID:   janeID,
			IsApproved:  true,
			CreatedAt:   time.Now().Add(-25 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-25 * 24 * time.Hour),
		},
	}

	if _, found := userMap["techguru"]; !found {
		communities[0].CreatorID = adminID
	}

	if _, found := userMap["fitnesscoach"]; !found {
		communities[1].CreatorID = janeID
	}

	if err := s.db.Table("communities").Create(&communities).Error; err != nil {
		return fmt.Errorf("failed to create communities: %w", err)
	}

	log.Printf("Created %d communities", len(communities))
	return nil
}

func (s *CommunitySeeder) SeedCommunityMembers() error {
	var count int64
	s.db.Table("community_members").Count(&count)
	if count > 0 {
		log.Println("Community members already exist, skipping seeding")
		return nil
	}

	adminID := uuid.MustParse("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11")
	johnID := uuid.MustParse("b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12")
	janeID := uuid.MustParse("c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a13")
	samID := uuid.MustParse("d0eebc99-9c0b-4ef8-bb6d-6bb9bd380a14")

	var communities []struct {
		CommunityID uuid.UUID
		Name        string
		CreatorID   uuid.UUID
	}
	s.db.Table("communities").Select("community_id, name, creator_id").Find(&communities)

	communityMap := make(map[string]uuid.UUID)
	creatorMap := make(map[uuid.UUID]uuid.UUID) 
	for _, community := range communities {
		communityMap[community.Name] = community.CommunityID
		creatorMap[community.CommunityID] = community.CreatorID
	}

	type CommunityMember struct {
		ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
		CommunityID uuid.UUID `gorm:"type:uuid;not null"`
		UserID      uuid.UUID `gorm:"type:uuid;not null"`
		Role        string    `gorm:"type:varchar(50);not null"`
		JoinedAt    time.Time
		CreatedAt   time.Time
		UpdatedAt   time.Time
	}

	members := []CommunityMember{}

	for communityID, creatorID := range creatorMap {
		members = append(members, CommunityMember{
			ID:          uuid.New(),
			CommunityID: communityID,
			UserID:      creatorID,
			Role:        "admin",
			JoinedAt:    time.Now().Add(-60 * 24 * time.Hour),
			CreatedAt:   time.Now().Add(-60 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-60 * 24 * time.Hour),
		})
	}

	if techID, ok := communityMap["Tech Enthusiasts"]; ok {
		members = append(members,
			CommunityMember{
				ID:          uuid.New(),
				CommunityID: techID,
				UserID:      johnID,
				Role:        "member",
				JoinedAt:    time.Now().Add(-59 * 24 * time.Hour),
				CreatedAt:   time.Now().Add(-59 * 24 * time.Hour),
				UpdatedAt:   time.Now().Add(-59 * 24 * time.Hour),
			},
			CommunityMember{
				ID:          uuid.New(),
				CommunityID: techID,
				UserID:      janeID,
				Role:        "moderator",
				JoinedAt:    time.Now().Add(-58 * 24 * time.Hour),
				CreatedAt:   time.Now().Add(-58 * 24 * time.Hour),
				UpdatedAt:   time.Now().Add(-58 * 24 * time.Hour),
			},
		)
	}

	if fitnessID, ok := communityMap["Fitness & Health"]; ok {
		members = append(members,
			CommunityMember{
				ID:          uuid.New(),
				CommunityID: fitnessID,
				UserID:      adminID,
				Role:        "member",
				JoinedAt:    time.Now().Add(-54 * 24 * time.Hour),
				CreatedAt:   time.Now().Add(-54 * 24 * time.Hour),
				UpdatedAt:   time.Now().Add(-54 * 24 * time.Hour),
			},
			CommunityMember{
				ID:          uuid.New(),
				CommunityID: fitnessID,
				UserID:      samID,
				Role:        "member",
				JoinedAt:    time.Now().Add(-53 * 24 * time.Hour),
				CreatedAt:   time.Now().Add(-53 * 24 * time.Hour),
				UpdatedAt:   time.Now().Add(-53 * 24 * time.Hour),
			},
		)
	}

	if devID, ok := communityMap["Developers Hub"]; ok {
		members = append(members,
			CommunityMember{
				ID:          uuid.New(),
				CommunityID: devID,
				UserID:      adminID,
				Role:        "moderator",
				JoinedAt:    time.Now().Add(-49 * 24 * time.Hour),
				CreatedAt:   time.Now().Add(-49 * 24 * time.Hour),
				UpdatedAt:   time.Now().Add(-49 * 24 * time.Hour),
			},
			CommunityMember{
				ID:          uuid.New(),
				CommunityID: devID,
				UserID:      janeID,
				Role:        "member",
				JoinedAt:    time.Now().Add(-48 * 24 * time.Hour),
				CreatedAt:   time.Now().Add(-48 * 24 * time.Hour),
				UpdatedAt:   time.Now().Add(-48 * 24 * time.Hour),
			},
		)
	}

	for _, community := range communities {

		if community.Name == "Tech Enthusiasts" || community.Name == "Fitness & Health" || community.Name == "Developers Hub" {
			continue
		}

		members = append(members,
			CommunityMember{
				ID:          uuid.New(),
				CommunityID: community.CommunityID,
				UserID:      adminID,
				Role:        "member",
				JoinedAt:    time.Now().Add(-40 * 24 * time.Hour),
				CreatedAt:   time.Now().Add(-40 * 24 * time.Hour),
				UpdatedAt:   time.Now().Add(-40 * 24 * time.Hour),
			},
		)

		if community.CreatorID != johnID {
			members = append(members,
				CommunityMember{
					ID:          uuid.New(),
					CommunityID: community.CommunityID,
					UserID:      johnID,
					Role:        "member",
					JoinedAt:    time.Now().Add(-39 * 24 * time.Hour),
					CreatedAt:   time.Now().Add(-39 * 24 * time.Hour),
					UpdatedAt:   time.Now().Add(-39 * 24 * time.Hour),
				},
			)
		}
	}

	if err := s.db.Table("community_members").Create(&members).Error; err != nil {
		return fmt.Errorf("failed to create community members: %w", err)
	}

	log.Printf("Created %d community members", len(members))
	return nil
}

func (s *CommunitySeeder) SeedJoinRequests() error {
	var count int64
	s.db.Table("community_join_requests").Count(&count)
	if count > 0 {
		log.Println("Community join requests already exist, skipping seeding")
		return nil
	}

	adminID := uuid.MustParse("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11")
	johnID := uuid.MustParse("b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12")
	janeID := uuid.MustParse("c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a13")
	samID := uuid.MustParse("d0eebc99-9c0b-4ef8-bb6d-6bb9bd380a14")

	var communities []struct {
		CommunityID uuid.UUID
		Name        string
	}
	s.db.Table("communities").Select("community_id, name").Find(&communities)

	communityMap := make(map[string]uuid.UUID)
	for _, community := range communities {
		communityMap[community.Name] = community.CommunityID
	}

	type CommunityJoinRequest struct {
		ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
		CommunityID uuid.UUID `gorm:"type:uuid;not null"`
		UserID      uuid.UUID `gorm:"type:uuid;not null"`
		Status      string    `gorm:"type:varchar(50);not null"`
		Message     string    `gorm:"type:text"`
		CreatedAt   time.Time
		UpdatedAt   time.Time
	}

	requests := []CommunityJoinRequest{}

	if techID, ok := communityMap["Tech Enthusiasts"]; ok {
		requests = append(requests, CommunityJoinRequest{
			ID:          uuid.New(),
			CommunityID: techID,
			UserID:      samID,
			Status:      "pending",
			Message:     "I'm really interested in technology and would love to join your community!",
			CreatedAt:   time.Now().Add(-10 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-10 * 24 * time.Hour),
		})
	}

	if bookID, ok := communityMap["Book Club"]; ok {
		requests = append(requests, CommunityJoinRequest{
			ID:          uuid.New(),
			CommunityID: bookID,
			UserID:      janeID,
			Status:      "pending",
			Message:     "I'm an avid reader and would like to join your book discussions.",
			CreatedAt:   time.Now().Add(-8 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-8 * 24 * time.Hour),
		})
	}

	if photoID, ok := communityMap["Photography Club"]; ok {
		requests = append(requests, CommunityJoinRequest{
			ID:          uuid.New(),
			CommunityID: photoID,
			UserID:      adminID,
			Status:      "pending",
			Message:     "Photography is one of my hobbies. I'd love to share my work and learn from others.",
			CreatedAt:   time.Now().Add(-6 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-6 * 24 * time.Hour),
		})
	}

	if designID, ok := communityMap["Design Showcase"]; ok {
		requests = append(requests, CommunityJoinRequest{
			ID:          uuid.New(),
			CommunityID: designID,
			UserID:      johnID,
			Status:      "pending",
			Message:     "I'm a developer looking to improve my design skills. Would love to join!",
			CreatedAt:   time.Now().Add(-5 * 24 * time.Hour),
			UpdatedAt:   time.Now().Add(-5 * 24 * time.Hour),
		})
	}

	if err := s.db.Table("community_join_requests").Create(&requests).Error; err != nil {
		return fmt.Errorf("failed to create community join requests: %w", err)
	}

	log.Printf("Created %d community join requests", len(requests))
	return nil
}