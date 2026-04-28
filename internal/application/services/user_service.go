// internal/application/services/user_service.go
package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	userdto "github.com/franciscozamorau/osmi-server/internal/api/dto/user"
	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
	"github.com/franciscozamorau/osmi-server/internal/domain/repository"
	"github.com/franciscozamorau/osmi-server/internal/infrastructure/cache"
	"github.com/franciscozamorau/osmi-server/internal/shared/security"
	"github.com/google/uuid"
)

type UserService struct {
	userRepo     repository.UserRepository
	customerRepo repository.CustomerRepository
	sessionRepo  repository.SessionRepository
	hasher       *security.PasswordHasher
	jwtService   *security.JWTService
	redisClient  *cache.RedisClient
}

func NewUserService(
	userRepo repository.UserRepository,
	customerRepo repository.CustomerRepository,
	sessionRepo repository.SessionRepository,
	hasher *security.PasswordHasher,
	jwtService *security.JWTService,
	redisClient *cache.RedisClient,
) *UserService {
	return &UserService{
		userRepo:     userRepo,
		customerRepo: customerRepo,
		sessionRepo:  sessionRepo,
		hasher:       hasher,
		jwtService:   jwtService,
		redisClient:  redisClient,
	}
}

// Register registra un nuevo usuario
func (s *UserService) Register(ctx context.Context, req *userdto.CreateUserRequest) (*entities.User, error) {
	if err := s.validateCreateUserRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	existing, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existing != nil {
		return nil, errors.New("email already registered")
	}
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}

	passwordHash, err := s.hasher.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	now := time.Now()

	var phone *string
	if req.Phone != "" && req.Phone != "null" {
		phone = &req.Phone
	}

	username := req.Username
	if username == "" {
		username = req.Email
	}

	user := &entities.User{
		PublicID:          uuid.New().String(),
		Email:             req.Email,
		Phone:             phone,
		Username:          &username,
		PasswordHash:      passwordHash,
		IsActive:          true,
		EmailVerified:     false,
		PhoneVerified:     false,
		PreferredLanguage: "es",
		PreferredCurrency: "MXN",
		Timezone:          "UTC",
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if req.FirstName != "" {
		user.FirstName = &req.FirstName
	}
	if req.LastName != "" {
		user.LastName = &req.LastName
	}
	if req.FirstName != "" && req.LastName != "" {
		fullName := req.FirstName + " " + req.LastName
		user.FullName = &fullName
	}

	role := req.Role
	if role == "" {
		role = "customer"
	}
	user.SetRole(role)

	if req.DateOfBirth != "" {
		dob, err := time.Parse("2006-01-02", req.DateOfBirth)
		if err == nil {
			user.DateOfBirth = &dob
		}
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	customer := &entities.Customer{
		PublicID:  uuid.New().String(),
		UserID:    &user.ID,
		FullName:  user.GetDisplayName(),
		Email:     user.Email,
		Phone:     phone,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.customerRepo.Create(ctx, customer); err != nil {
		log.Printf("Warning: failed to create customer profile for user %s: %v", user.PublicID, err)
	}

	return user, nil
}

// AuthResponse es la estructura que devuelve autenticación
type AuthResponse struct {
	PublicID  string
	Email     string
	Username  *string
	Role      string
	CreatedAt time.Time
}

// Authenticate verifica credenciales y devuelve el usuario autenticado
func (s *UserService) Authenticate(ctx context.Context, email, password string) (*AuthResponse, error) {
	log.Printf("🔐 Authenticate llamado con email: %s, password: %s", email, password)

	if email == "" || password == "" {
		return nil, errors.New("email and password are required")
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, errors.New("invalid credentials")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		return nil, errors.New("account is locked")
	}

	if !user.IsActive {
		return nil, errors.New("account is inactive")
	}

	if !s.hasher.VerifyPassword(user.PasswordHash, password) {
		user.FailedLoginAttempts++
		user.UpdatedAt = time.Now()
		_ = s.userRepo.Update(ctx, user)
		return nil, errors.New("invalid credentials")
	}

	user.FailedLoginAttempts = 0
	user.UpdatedAt = time.Now()
	_ = s.userRepo.Update(ctx, user)
	_ = s.userRepo.UpdateLastLogin(ctx, user.ID, "")

	role := "customer"
	if user.IsSuperuser {
		role = "admin"
	} else if user.IsStaff {
		role = "staff"
	}

	return &AuthResponse{
		PublicID:  user.PublicID,
		Email:     user.Email,
		Username:  user.Username,
		Role:      role,
		CreatedAt: user.CreatedAt,
	}, nil
}

// GetProfile obtiene el perfil de un usuario
func (s *UserService) GetProfile(ctx context.Context, userID int64) (*entities.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

// UpdateProfile actualiza el perfil de un usuario
func (s *UserService) UpdateProfile(ctx context.Context, userID int64, req *userdto.UpdateUserRequest) (*entities.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if req.FirstName != nil {
		user.FirstName = req.FirstName
	}
	if req.LastName != nil {
		user.LastName = req.LastName
	}
	if req.FirstName != nil && req.LastName != nil {
		fullName := *req.FirstName + " " + *req.LastName
		user.FullName = &fullName
	}
	if req.Phone != nil {
		user.Phone = req.Phone
	}
	if req.AvatarURL != nil {
		user.AvatarURL = req.AvatarURL
	}
	if req.DateOfBirth != nil {
		dob, err := time.Parse("2006-01-02", *req.DateOfBirth)
		if err == nil {
			user.DateOfBirth = &dob
		}
	}
	if req.PreferredLanguage != nil {
		user.PreferredLanguage = *req.PreferredLanguage
	}
	if req.PreferredCurrency != nil {
		user.PreferredCurrency = *req.PreferredCurrency
	}
	if req.Timezone != nil {
		user.Timezone = *req.Timezone
	}

	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// ChangePassword cambia la contraseña de un usuario
func (s *UserService) ChangePassword(ctx context.Context, userID int64, req *userdto.ChangePasswordRequest) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return errors.New("user not found")
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	if !s.hasher.VerifyPassword(user.PasswordHash, req.CurrentPassword) {
		return errors.New("current password is incorrect")
	}

	newHash, err := s.hasher.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	user.PasswordHash = newHash
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// Logout invalida un token (lo agrega a blacklist en Redis)
func (s *UserService) Logout(ctx context.Context, token string) error {
	if token == "" {
		return errors.New("token is required")
	}

	claims, err := s.jwtService.ValidateToken(token)
	if err != nil {
		return errors.New("invalid token")
	}

	expiresAt := claims.ExpiresAt.Time
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return errors.New("token already expired")
	}

	if err := s.redisClient.AddToBlacklist(ctx, token, ttl); err != nil {
		return fmt.Errorf("failed to blacklist token: %w", err)
	}

	log.Printf("🔐 Token blacklisted para user_id: %s, expira en: %v", claims.UserID, ttl)
	return nil
}

// RefreshToken genera un nuevo token
func (s *UserService) RefreshToken(ctx context.Context, oldToken string) (string, time.Time, error) {
	claims, err := s.jwtService.ValidateToken(oldToken)
	if err != nil {
		return "", time.Time{}, errors.New("invalid token")
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	newToken, err := s.jwtService.GenerateAccessToken(claims.UserID)
	if err != nil {
		return "", time.Time{}, err
	}

	return newToken, expiresAt, nil
}

// LogoutAll cierra todas las sesiones de un usuario
func (s *UserService) LogoutAll(ctx context.Context, userID int64) error {
	if err := s.sessionRepo.InvalidateAllForUser(ctx, userID); err != nil {
		return fmt.Errorf("failed to logout all sessions: %w", err)
	}
	return nil
}

// DeleteAccount desactiva la cuenta de un usuario
func (s *UserService) DeleteAccount(ctx context.Context, userID int64) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return errors.New("user not found")
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	user.IsActive = false
	user.UpdatedAt = time.Now()
	_ = s.sessionRepo.InvalidateAllForUser(ctx, userID)

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to delete account: %w", err)
	}

	return nil
}

// validateCreateUserRequest valida los datos de registro
func (s *UserService) validateCreateUserRequest(req *userdto.CreateUserRequest) error {
	if req.Email == "" {
		return errors.New("email is required")
	}
	if req.Password == "" {
		return errors.New("password is required")
	}
	if len(req.Password) < 6 {
		return errors.New("password must be at least 6 characters")
	}
	if req.Username == "" {
		return errors.New("username is required")
	}
	if len(req.Username) < 3 {
		return errors.New("username must be at least 3 characters")
	}
	return nil
}

// GetUserByPublicID obtiene un usuario por su PublicID (UUID)
func (s *UserService) GetUserByPublicID(ctx context.Context, publicID string) (*entities.User, error) {
	user, err := s.userRepo.GetByPublicID(ctx, publicID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

// ListUsers lista todos los usuarios activos
func (s *UserService) ListUsers(ctx context.Context, page, pageSize int) ([]*entities.User, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	return s.userRepo.List(ctx, pageSize, offset)
}

// UpdateUser actualiza un usuario existente
func (s *UserService) UpdateUser(ctx context.Context, publicID string, req *userdto.UpdateUserRequest) (*entities.User, error) {
	log.Printf("📝 UpdateUser service: publicID=%s", publicID)

	user, err := s.userRepo.GetByPublicID(ctx, publicID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if req.FirstName != nil && *req.FirstName != "" {
		log.Printf("📝 Actualizando username de '%s' a '%s'", *user.Username, *req.FirstName)
		user.Username = req.FirstName
	}

	if req.Phone != nil {
		user.Phone = req.Phone
	}
	if req.AvatarURL != nil {
		user.AvatarURL = req.AvatarURL
	}
	if req.PreferredLanguage != nil {
		user.PreferredLanguage = *req.PreferredLanguage
	}
	if req.PreferredCurrency != nil {
		user.PreferredCurrency = *req.PreferredCurrency
	}
	if req.Timezone != nil {
		user.Timezone = *req.Timezone
	}

	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	log.Printf("✅ Usuario actualizado correctamente")
	return user, nil
}
