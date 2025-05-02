package dto

type ReplyDetailDTO struct {
	ReplyID   string `json:"reply_id"`
	ThreadID  string `json:"thread_id"`
	UserID    string `json:"user_id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
