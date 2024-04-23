package customer

type SignUpRequest struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"email"`
	Password string `json:"password" validate:"required"`
}

type SignInRequest struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type VerifyRequest struct {
	Token string
}

type UpdateProfileRequest struct {
	Name string `json:"name"`
}

type ChangeEmailRequest struct {
	Email string `json:"email" validate:"email"`
}

type ChangePasswordRequest struct {
	ExistingPassword string `json:"existing_password" validate:"required"`
	NewPassword      string `json:"new_password" validate:"required"`
}

type ChangeEmailVerificationRequest struct {
	Token string
}
