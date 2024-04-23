package customer

import "time"

type SignUpResponse struct {
	VerificationExpiresAt time.Time `json:"verification_expires_at"`
}

type SignInResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

type GetProfileResponse struct {
	ID                 int64     `json:"id"`
	Name               string    `json:"name"`
	Email              string    `json:"email"`
	VerificationStatus string    `json:"verification_status"`
	MemberStatus       string    `json:"member_status"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type ChangeEmailResponse struct {
	VerificationExpiresAt time.Time `json:"verification_expires_at"`
}
