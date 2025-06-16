package db

import (
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"aycom/backend/services/thread/model"
)

func CheckUserDataExists(db *gorm.DB) bool {
	adminID := uuid.MustParse("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11")

	var count int64
	db.Table("users").Where("id = ?", adminID).Count(&count)

	if count == 0 {
		log.Println("Warning: User data not found. Make sure to run user migrations and seeders first.")
		return false
	}

	return true
}

func SeedDatabase(db *gorm.DB) error {
	log.Println("Starting thread database seeding...")

	if err := seedCategories(db); err != nil {
		return err
	}

	var count int64
	db.Model(&model.Thread{}).Count(&count)
	if count > 0 {
		log.Println("Threads already exist, skipping seeding")
		return nil
	}

	if !CheckUserDataExists(db) {
		log.Println("Unable to seed thread data: user data doesn't exist yet")
		return nil
	}

	categoryMap := make(map[string]uuid.UUID)
	var categories []model.Category
	db.Find(&categories)
	for _, cat := range categories {
		categoryMap[cat.Name] = cat.CategoryID
	}

	adminID := uuid.MustParse("a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11")
	johnID := uuid.MustParse("b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12")
	janeID := uuid.MustParse("c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a13")
	samID := uuid.MustParse("d0eebc99-9c0b-4ef8-bb6d-6bb9bd380a14")

	var additionalUsers []struct {
		ID       uuid.UUID
		Username string
	}
	db.Table("users").Select("id, username").Where("username IN ?", []string{"techguru", "fitnesscoach", "travelbug", "foodie123"}).Find(&additionalUsers)

	userMap := make(map[string]uuid.UUID)
	for _, user := range additionalUsers {
		userMap[user.Username] = user.ID
	}

	threads := []model.Thread{

		{
			ThreadID:  uuid.New(),
			UserID:    adminID,
			Content:   "Welcome to AYCOM! This is our first official announcement. #welcome #aycom",
			IsPinned:  true,
			CreatedAt: time.Now().Add(-48 * time.Hour),
			UpdatedAt: time.Now().Add(-48 * time.Hour),
		},
		{
			ThreadID:  uuid.New(),
			UserID:    adminID,
			Content:   "We're excited to announce new features coming soon! #update #features",
			CreatedAt: time.Now().Add(-24 * time.Hour),
			UpdatedAt: time.Now().Add(-24 * time.Hour),
		},

		{
			ThreadID:  uuid.New(),
			UserID:    johnID,
			Content:   "Hello everyone! Just joined this platform and it looks amazing. #newuser #hello",
			CreatedAt: time.Now().Add(-36 * time.Hour),
			UpdatedAt: time.Now().Add(-36 * time.Hour),
		},
		{
			ThreadID:  uuid.New(),
			UserID:    johnID,
			Content:   "What's everyone working on today? #coding #programming #webdev",
			CreatedAt: time.Now().Add(-12 * time.Hour),
			UpdatedAt: time.Now().Add(-12 * time.Hour),
		},

		{
			ThreadID:  uuid.New(),
			UserID:    janeID,
			Content:   "Just finished my new project. Check it out! #project #design #portfolio",
			CreatedAt: time.Now().Add(-18 * time.Hour),
			UpdatedAt: time.Now().Add(-18 * time.Hour),
		},
		{
			ThreadID:  uuid.New(),
			UserID:    janeID,
			Content:   "Looking for collaborators on a new open source project. DM me if interested. #opensource #collaboration",
			CreatedAt: time.Now().Add(-6 * time.Hour),
			UpdatedAt: time.Now().Add(-6 * time.Hour),
		},

		{
			ThreadID:  uuid.New(),
			UserID:    samID,
			Content:   "First day at my new job! Excited for this new chapter. #newjob #career",
			CreatedAt: time.Now().Add(-30 * time.Hour),
			UpdatedAt: time.Now().Add(-30 * time.Hour),
		},
		{
			ThreadID:  uuid.New(),
			UserID:    samID,
			Content:   "Anyone else attending the tech conference next week? #techconf #networking",
			CreatedAt: time.Now().Add(-8 * time.Hour),
			UpdatedAt: time.Now().Add(-8 * time.Hour),
		},
	}

	if techGuruID, ok := userMap["techguru"]; ok {
		threads = append(threads,
			model.Thread{
				ThreadID:  uuid.New(),
				UserID:    techGuruID,
				Content:   "Breaking news: New AI breakthrough announced today! Just saw the demo and it's mind-blowing. #AI #innovation #technology",
				CreatedAt: time.Now().Add(-5 * time.Hour),
				UpdatedAt: time.Now().Add(-5 * time.Hour),
			},
			model.Thread{
				ThreadID:  uuid.New(),
				UserID:    techGuruID,
				Content:   "My review of the latest smartphone just dropped. Check it out and let me know your thoughts! #review #smartphone #tech",
				CreatedAt: time.Now().Add(-3 * time.Hour),
				UpdatedAt: time.Now().Add(-3 * time.Hour),
			},
		)
	}

	if fitnessCoachID, ok := userMap["fitnesscoach"]; ok {
		threads = append(threads,
			model.Thread{
				ThreadID:  uuid.New(),
				UserID:    fitnessCoachID,
				Content:   "Morning workout complete! Starting the day with energy and focus. #fitness #health #wellness",
				CreatedAt: time.Now().Add(-7 * time.Hour),
				UpdatedAt: time.Now().Add(-7 * time.Hour),
			},
			model.Thread{
				ThreadID:  uuid.New(),
				UserID:    fitnessCoachID,
				Content:   "New workout routine for beginners! Sharing my top 5 exercises for those just starting their fitness journey. #fitness #beginners #workout",
				CreatedAt: time.Now().Add(-2 * time.Hour),
				UpdatedAt: time.Now().Add(-2 * time.Hour),
			},
		)
	}

	if travelBugID, ok := userMap["travelbug"]; ok {
		threads = append(threads,
			model.Thread{
				ThreadID:  uuid.New(),
				UserID:    travelBugID,
				Content:   "Just booked my next adventure to Bali! Can't wait to explore this beautiful island. #travel #bali #vacation",
				CreatedAt: time.Now().Add(-10 * time.Hour),
				UpdatedAt: time.Now().Add(-10 * time.Hour),
			},
			model.Thread{
				ThreadID:  uuid.New(),
				UserID:    travelBugID,
				Content:   "Travel tip: Always pack a portable charger and universal adapter. They've saved me countless times! #travel #traveltips",
				CreatedAt: time.Now().Add(-4 * time.Hour),
				UpdatedAt: time.Now().Add(-4 * time.Hour),
			},
		)
	}

	if foodieID, ok := userMap["foodie123"]; ok {
		threads = append(threads,
			model.Thread{
				ThreadID:  uuid.New(),
				UserID:    foodieID,
				Content:   "Visited this amazing restaurant yesterday. The food was incredible! #food #foodie #restaurant",
				CreatedAt: time.Now().Add(-9 * time.Hour),
				UpdatedAt: time.Now().Add(-9 * time.Hour),
			},
			model.Thread{
				ThreadID:  uuid.New(),
				UserID:    foodieID,
				Content:   "Recipe of the day: Quick and easy pasta carbonara. Ready in 15 minutes! #food #recipe #cooking",
				CreatedAt: time.Now().Add(-1 * time.Hour),
				UpdatedAt: time.Now().Add(-1 * time.Hour),
			},
		)
	}

	result := db.Create(&threads)
	if result.Error != nil {
		log.Printf("Error seeding threads: %v", result.Error)
		return result.Error
	}

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
			Text:      "webdev",
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
		{
			HashtagID: uuid.New(),
			Text:      "AI",
			CreatedAt: time.Now(),
		},
		{
			HashtagID: uuid.New(),
			Text:      "innovation",
			CreatedAt: time.Now(),
		},
		{
			HashtagID: uuid.New(),
			Text:      "technology",
			CreatedAt: time.Now(),
		},
		{
			HashtagID: uuid.New(),
			Text:      "design",
			CreatedAt: time.Now(),
		},
		{
			HashtagID: uuid.New(),
			Text:      "portfolio",
			CreatedAt: time.Now(),
		},
		{
			HashtagID: uuid.New(),
			Text:      "fitness",
			CreatedAt: time.Now(),
		},
		{
			HashtagID: uuid.New(),
			Text:      "health",
			CreatedAt: time.Now(),
		},
		{
			HashtagID: uuid.New(),
			Text:      "wellness",
			CreatedAt: time.Now(),
		},
		{
			HashtagID: uuid.New(),
			Text:      "travel",
			CreatedAt: time.Now(),
		},
		{
			HashtagID: uuid.New(),
			Text:      "bali",
			CreatedAt: time.Now(),
		},
		{
			HashtagID: uuid.New(),
			Text:      "vacation",
			CreatedAt: time.Now(),
		},
		{
			HashtagID: uuid.New(),
			Text:      "food",
			CreatedAt: time.Now(),
		},
		{
			HashtagID: uuid.New(),
			Text:      "foodie",
			CreatedAt: time.Now(),
		},
		{
			HashtagID: uuid.New(),
			Text:      "recipe",
			CreatedAt: time.Now(),
		},
		{
			HashtagID: uuid.New(),
			Text:      "cooking",
			CreatedAt: time.Now(),
		},
	}

	result = db.Create(&hashtags)
	if result.Error != nil {
		log.Printf("Error seeding hashtags: %v", result.Error)
		return result.Error
	}

	hashtagMap := make(map[string]uuid.UUID)
	for _, hashtag := range hashtags {
		hashtagMap[hashtag.Text] = hashtag.HashtagID
	}

	var threadHashtags []model.ThreadHashtag

	extractAndCreateHashtags := func(thread model.Thread) {

		for tag, id := range hashtagMap {
			if tag == "coding" && thread.ThreadID == threads[3].ThreadID {
				threadHashtags = append(threadHashtags, model.ThreadHashtag{
					ThreadID:  thread.ThreadID,
					HashtagID: id,
				})
			}
			if tag == "programming" && thread.ThreadID == threads[3].ThreadID {
				threadHashtags = append(threadHashtags, model.ThreadHashtag{
					ThreadID:  thread.ThreadID,
					HashtagID: id,
				})
			}
			if tag == "webdev" && thread.ThreadID == threads[3].ThreadID {
				threadHashtags = append(threadHashtags, model.ThreadHashtag{
					ThreadID:  thread.ThreadID,
					HashtagID: id,
				})
			}

		}
	}

	for _, thread := range threads {
		extractAndCreateHashtags(thread)
	}

	threadHashtags = []model.ThreadHashtag{

		{
			ThreadID:  threads[3].ThreadID,
			HashtagID: hashtagMap["coding"],
		},
		{
			ThreadID:  threads[3].ThreadID,
			HashtagID: hashtagMap["programming"],
		},
		{
			ThreadID:  threads[3].ThreadID,
			HashtagID: hashtagMap["webdev"],
		},

		{
			ThreadID:  threads[4].ThreadID,
			HashtagID: hashtagMap["design"],
		},
		{
			ThreadID:  threads[4].ThreadID,
			HashtagID: hashtagMap["portfolio"],
		},

		{
			ThreadID:  threads[5].ThreadID,
			HashtagID: hashtagMap["opensource"],
		},

		{
			ThreadID:  threads[7].ThreadID,
			HashtagID: hashtagMap["techconf"],
		},
	}

	for _, thread := range threads[8:] {

		if strings.Contains(thread.Content, "AI") {
			threadHashtags = append(threadHashtags, model.ThreadHashtag{
				ThreadID:  thread.ThreadID,
				HashtagID: hashtagMap["AI"],
			})
		}
		if strings.Contains(thread.Content, "innovation") {
			threadHashtags = append(threadHashtags, model.ThreadHashtag{
				ThreadID:  thread.ThreadID,
				HashtagID: hashtagMap["innovation"],
			})
		}
		if strings.Contains(thread.Content, "technology") {
			threadHashtags = append(threadHashtags, model.ThreadHashtag{
				ThreadID:  thread.ThreadID,
				HashtagID: hashtagMap["technology"],
			})
		}

	}

	result = db.Create(&threadHashtags)
	if result.Error != nil {
		log.Printf("Error seeding thread hashtags: %v", result.Error)
		return result.Error
	}

	media := []model.Media{}

	result = db.Create(&media)
	if result.Error != nil {
		log.Printf("Error seeding media: %v", result.Error)
		return result.Error
	}

	replies := []model.Reply{
		{
			ReplyID:   uuid.New(),
			ThreadID:  threads[0].ThreadID,
			UserID:    johnID,
			Content:   "Excited to be here!",
			CreatedAt: time.Now().Add(-47 * time.Hour),
			UpdatedAt: time.Now().Add(-47 * time.Hour),
		},
		{
			ReplyID:   uuid.New(),
			ThreadID:  threads[2].ThreadID,
			UserID:    janeID,
			Content:   "Welcome John! Nice to meet you.",
			CreatedAt: time.Now().Add(-35 * time.Hour),
			UpdatedAt: time.Now().Add(-35 * time.Hour),
		},
		{
			ReplyID:   uuid.New(),
			ThreadID:  threads[4].ThreadID,
			UserID:    johnID,
			Content:   "That looks amazing! Great work.",
			CreatedAt: time.Now().Add(-17 * time.Hour),
			UpdatedAt: time.Now().Add(-17 * time.Hour),
		},
		{
			ReplyID:   uuid.New(),
			ThreadID:  threads[6].ThreadID,
			UserID:    adminID,
			Content:   "Congratulations on the new job!",
			CreatedAt: time.Now().Add(-29 * time.Hour),
			UpdatedAt: time.Now().Add(-29 * time.Hour),
		},
	}

	result = db.Create(&replies)
	if result.Error != nil {
		log.Printf("Error seeding replies: %v", result.Error)
		return result.Error
	}

	likes := []model.Like{

		{
			UserID:    johnID,
			ThreadID:  &threads[0].ThreadID,
			CreatedAt: time.Now().Add(-46 * time.Hour),
		},
		{
			UserID:    janeID,
			ThreadID:  &threads[0].ThreadID,
			CreatedAt: time.Now().Add(-45 * time.Hour),
		},
		{
			UserID:    samID,
			ThreadID:  &threads[0].ThreadID,
			CreatedAt: time.Now().Add(-44 * time.Hour),
		},

		{
			UserID:    adminID,
			ThreadID:  &threads[2].ThreadID,
			CreatedAt: time.Now().Add(-34 * time.Hour),
		},
		{
			UserID:    janeID,
			ThreadID:  &threads[2].ThreadID,
			CreatedAt: time.Now().Add(-33 * time.Hour),
		},
		{
			UserID:    janeID,
			ThreadID:  &threads[3].ThreadID,
			CreatedAt: time.Now().Add(-11 * time.Hour),
		},
		{
			UserID:    adminID,
			ThreadID:  &threads[3].ThreadID,
			CreatedAt: time.Now().Add(-10 * time.Hour),
		},

		{
			UserID:    johnID,
			ThreadID:  &threads[4].ThreadID,
			CreatedAt: time.Now().Add(-17 * time.Hour),
		},
		{
			UserID:    adminID,
			ThreadID:  &threads[4].ThreadID,
			CreatedAt: time.Now().Add(-16 * time.Hour),
		},
		{
			UserID:    samID,
			ThreadID:  &threads[5].ThreadID,
			CreatedAt: time.Now().Add(-5 * time.Hour),
		},

		{
			UserID:    janeID,
			ThreadID:  &threads[6].ThreadID,
			CreatedAt: time.Now().Add(-28 * time.Hour),
		},
		{
			UserID:    johnID,
			ThreadID:  &threads[6].ThreadID,
			CreatedAt: time.Now().Add(-27 * time.Hour),
		},
	}

	for i := 8; i < len(threads); i++ {

		likes = append(likes,
			model.Like{
				UserID:    adminID,
				ThreadID:  &threads[i].ThreadID,
				CreatedAt: threads[i].CreatedAt.Add(1 * time.Hour),
			},
			model.Like{
				UserID:    johnID,
				ThreadID:  &threads[i].ThreadID,
				CreatedAt: threads[i].CreatedAt.Add(2 * time.Hour),
			},
			model.Like{
				UserID:    janeID,
				ThreadID:  &threads[i].ThreadID,
				CreatedAt: threads[i].CreatedAt.Add(3 * time.Hour),
			},
		)
	}

	result = db.Create(&likes)
	if result.Error != nil {
		log.Printf("Error seeding likes: %v", result.Error)
		return result.Error
	}

	log.Printf("Successfully seeded %d threads, %d hashtags, %d thread-hashtag associations, %d media items, %d replies, and %d likes",
		len(threads), len(hashtags), len(threadHashtags), len(media), len(replies), len(likes))
	return nil
}

func seedCategories(db *gorm.DB) error {
	var count int64
	db.Model(&model.Category{}).Count(&count)
	if count > 0 {
		log.Println("Categories already exist, skipping seeding")
		return nil
	}

	categories := []model.Category{
		{
			CategoryID: uuid.New(),
			Name:       "General",
			CreatedAt:  time.Now(),
		},
		{
			CategoryID: uuid.New(),
			Name:       "Technology",
			CreatedAt:  time.Now(),
		},
		{
			CategoryID: uuid.New(),
			Name:       "Health",
			CreatedAt:  time.Now(),
		},
		{
			CategoryID: uuid.New(),
			Name:       "Travel",
			CreatedAt:  time.Now(),
		},
		{
			CategoryID: uuid.New(),
			Name:       "Food",
			CreatedAt:  time.Now(),
		},
		{
			CategoryID: uuid.New(),
			Name:       "Design",
			CreatedAt:  time.Now(),
		},
		{
			CategoryID: uuid.New(),
			Name:       "Career",
			CreatedAt:  time.Now(),
		},
		{
			CategoryID: uuid.New(),
			Name:       "Announcements",
			CreatedAt:  time.Now(),
		},
		{
			CategoryID: uuid.New(),
			Name:       "Entertainment",
			CreatedAt:  time.Now(),
		},
		{
			CategoryID: uuid.New(),
			Name:       "Sports",
			CreatedAt:  time.Now(),
		},
		{
			CategoryID: uuid.New(),
			Name:       "Politics",
			CreatedAt:  time.Now(),
		},
		{
			CategoryID: uuid.New(),
			Name:       "Science",
			CreatedAt:  time.Now(),
		},
	}

	result := db.Create(&categories)
	if result.Error != nil {
		log.Printf("Error seeding categories: %v", result.Error)
		return result.Error
	}

	log.Printf("Successfully seeded %d categories", len(categories))
	return nil
}
