package dto

type CreateUserRequest struct {
	Email             string  `json:"email" validate:"required,email,max=320"`
	Phone             *string `json:"phone,omitempty" validate:"omitempty,phone"`
	Username          string  `json:"username,omitempty" validate:"omitempty,min=3,max=50,alphanum"`
	Password          string  `json:"password" validate:"required,min=8,strong_password"`
	FirstName         string  `json:"first_name,omitempty" validate:"omitempty,max=100,alpha"`
	LastName          string  `json:"last_name,omitempty" validate:"omitempty,max=100,alpha"`
	DateOfBirth       *string `json:"date_of_birth,omitempty" validate:"omitempty,date,valid_age"`
	PreferredLanguage string  `json:"preferred_language" validate:"required,oneof=es en pt"`
	PreferredCurrency string  `json:"preferred_currency" validate:"required,iso4217"`
	Timezone          string  `json:"timezone" validate:"required,timezone"`
	MarketingOptIn    *bool   `json:"marketing_opt_in" validate:"omitempty"`
	TermsAccepted     bool    `json:"terms_accepted" validate:"required,eq=true"`
}

type UpdateUserRequest struct {
	Phone             *string `json:"phone,omitempty" validate:"omitempty,phone"`
	FirstName         *string `json:"first_name,omitempty" validate:"omitempty,max=100,alpha"`
	LastName          *string `json:"last_name,omitempty" validate:"omitempty,max=100,alpha"`
	AvatarURL         *string `json:"avatar_url,omitempty" validate:"omitempty,url,max=500"`
	DateOfBirth       *string `json:"date_of_birth,omitempty" validate:"omitempty,date,valid_age"`
	PreferredLanguage *string `json:"preferred_language,omitempty" validate:"omitempty,oneof=es en pt"`
	PreferredCurrency *string `json:"preferred_currency,omitempty" validate:"omitempty,iso4217"`
	Timezone          *string `json:"timezone,omitempty" validate:"omitempty,timezone"`
	Bio               *string `json:"bio,omitempty" validate:"omitempty,max=500"`
	Website           *string `json:"website,omitempty" validate:"omitempty,url"`
	MarketingOptIn    *bool   `json:"marketing_opt_in,omitempty"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8,strong_password,nefield=CurrentPassword"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=NewPassword"`
}

type LoginRequest struct {
	Email      string  `json:"email" validate:"required,email"`
	Password   string  `json:"password" validate:"required"`
	DeviceID   *string `json:"device_id,omitempty" validate:"omitempty,max=255"`
	DeviceName *string `json:"device_name,omitempty" validate:"omitempty,max=100"`
	IPAddress  *string `json:"ip_address,omitempty" validate:"omitempty,ip"`
	RememberMe bool    `json:"remember_me"`
}

type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required,uuid4"`
	Email string `json:"email" validate:"required,email"`
}

type MFARequest struct {
	Code      string  `json:"code" validate:"required,len=6,numeric"`
	DeviceID  *string `json:"device_id,omitempty" validate:"omitempty,max=255"`
	SessionID string  `json:"session_id" validate:"required,uuid4"`
}

type ResetPasswordRequest struct {
	Token           string `json:"token" validate:"required,uuid4"`
	Email           string `json:"email" validate:"required,email"`
	NewPassword     string `json:"new_password" validate:"required,min=8,strong_password"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=NewPassword"`
}

type UpdateProfileRequest struct {
	FirstName *string `json:"first_name,omitempty" validate:"omitempty,max=100,alpha"`
	LastName  *string `json:"last_name,omitempty" validate:"omitempty,max=100,alpha"`
	AvatarURL *string `json:"avatar_url,omitempty" validate:"omitempty,url,max=500"`
	Bio       *string `json:"bio,omitempty" validate:"omitempty,max=500"`
	Website   *string `json:"website,omitempty" validate:"omitempty,url"`
}
