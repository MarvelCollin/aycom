package dto

type ThreadCreateDTO struct {
	UserID      string   `json:"user_id"`
	Content     string   `json:"content"`
	WhoCanReply string   `json:"who_can_reply"`
	MediaURLs   []string `json:"media_urls"`
	ScheduledAt string   `json:"scheduled_at"`
	CommunityID string   `json:"community_id"`
}
