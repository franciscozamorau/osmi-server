package request

type CreateUserRequest struct {
	Username          string `json:"username" validate:"required,min=3"`
	Email             string `json:"email" validate:"required,email"`
	Password          string `json:"password" validate:"required,min=6"`
	Phone             string `json:"phone,omitempty"`
	FirstName         string `json:"first_name"`
	LastName          string `json:"last_name"`
	DateOfBirth       string `json:"date_of_birth,omitempty" validate:"omitempty,datetime=2006-01-02"`
	PreferredLanguage string `json:"preferred_language" validate:"len=2"`
	PreferredCurrency string `json:"preferred_currency" validate:"len=3"`
	Timezone          string `json:"timezone"`
	Role              string `json:"role" validate:"oneof=admin customer organizer guest"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type UpdateUserRequest struct {
	Username          *string `json:"username,omitempty" validate:"omitempty,min=3"`
	Email             *string `json:"email,omitempty" validate:"omitempty,email"`
	Phone             *string `json:"phone,omitempty"`
	FirstName         *string `json:"first_name,omitempty"`
	LastName          *string `json:"last_name,omitempty"`
	DateOfBirth       *string `json:"date_of_birth,omitempty" validate:"omitempty,datetime=2006-01-02"`
	PreferredLanguage *string `json:"preferred_language,omitempty" validate:"omitempty,len=2"`
	PreferredCurrency *string `json:"preferred_currency,omitempty" validate:"omitempty,len=3"`
	Timezone          *string `json:"timezone,omitempty"`
	AvatarURL         *string `json:"avatar_url,omitempty"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=6"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=NewPassword"`
}
