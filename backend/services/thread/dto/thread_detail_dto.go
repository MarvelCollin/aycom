package dto

type ThreadDetailDTO struct {
	ThreadID        string   `json:"thread_id"`
	UserID          string   `json:"user_id"`
	Content         string   `json:"content"`
	IsPinned        bool     `json:"is_pinned"`
	WhoCanReply     string   `json:"who_can_reply"`
	ScheduledAt     string   `json:"scheduled_at"`
	CommunityID     string   `json:"community_id"`
	IsAdvertisement bool     `json:"is_advertisement"`
	CreatedAt       string   `json:"created_at"`
	UpdatedAt       string   `json:"updated_at"`
	Media           []string `json:"media"`
}
