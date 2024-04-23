package customer

import "time"

type ChangeEmailEvent struct {
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	ExistingEmail   string `json:"existing_email"`
	NewEmail        string `json:"new_email"`
	VerficationLink string `json:"verification_link"`
}

type SignUpEvent struct {
	ID                 int64     `json:"id"`
	Name               string    `json:"name"`
	Email              string    `json:"email"`
	VerificationStatus string    `json:"verification_status"`
	MemberStatus       string    `json:"member_status"`
	VerificationLink   string    `json:"verification_link"`
	CreatedAt          time.Time `json:"created_at"`
}
