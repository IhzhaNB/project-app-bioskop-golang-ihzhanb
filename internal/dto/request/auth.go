package request

type RegisterRequest struct {
	Username string  `json:"username" validate:"required,min=3,max=50"`
	Email    string  `json:"email" validate:"required,email"`
	Password string  `json:"password" validate:"required,min=6"`
	Phone    *string `json:"phone,omitempty" validate:"omitempty,min=10,max=15"`
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required,min=6"`
}

type VerifyEmailRequest struct {
	Email string `json:"email" validate:"required,email"`
	OTP   string `json:"otp" validate:"required,len=6"`
}

type SendOTPRequest struct {
	Email string `json:"email" validate:"required,email"`
	Type  string `json:"type" validate:"required,oneof=email_verification password_reset"`
}
