package dto

type UserCreateDTO struct {
	Name                  string `json:"name"`
	Username              string `json:"username"`
	Email                 string `json:"email"`
	Password              string `json:"password"`
	Gender                string `json:"gender"`
	DateOfBirth           string `json:"date_of_birth"`
	SecurityQuestion      string `json:"security_question"`
	SecurityAnswer        string `json:"security_answer"`
	SubscribeToNewsletter bool   `json:"subscribe_to_newsletter"`
}
