// internal/application/services/user_service.go
package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/franciscozamorau/osmi-server/internal/api/dto/request"
	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
	"github.com/franciscozamorau/osmi-server/internal/domain/enums"
	"github.com/franciscozamorau/osmi-server/internal/domain/repository"
	"github.com/franciscozamorau/osmi-server/internal/shared/security"
	"github.com/google/uuid"
)

type UserService struct {
	userRepo     repository.UserRepository
	customerRepo repository.CustomerRepository
	sessionRepo  repository.SessionRepository
	hasher       *security.PasswordHasher
	jwtService   *security.JWTService
}

func NewUserService(
	userRepo repository.UserRepository,
	customerRepo repository.CustomerRepository,
	sessionRepo repository.SessionRepository,
	hasher *security.PasswordHasher,
	jwtService *security.JWTService,
) *UserService {
	return &UserService{
		userRepo:     userRepo,
		customerRepo: customerRepo,
		sessionRepo:  sessionRepo,
		hasher:       hasher,
		jwtService:   jwtService,
	}
}

// Register registra un nuevo usuario en el sistema
func (s *UserService) Register(ctx context.Context, req *request.CreateUserRequest) (*entities.User, error) {
	// Validar request
	if err := s.validateCreateUserRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Verificar si el email ya existe
	existing, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existing != nil {
		return nil, errors.New("email already registered")
	}
	// Si el error es diferente a "not found", algo salió mal
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}

	// Hashear contraseña
	passwordHash, err := s.hasher.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Crear usuario
	now := time.Now()
	user := &entities.User{
		PublicID:          uuid.New().String(),
		Email:             req.Email,
		Phone:             &req.Phone,
		Username:          &req.Username,
		PasswordHash:      passwordHash,
		FirstName:         &req.FirstName,
		LastName:          &req.LastName,
		PreferredLanguage: req.PreferredLanguage,
		PreferredCurrency: req.PreferredCurrency,
		Timezone:          req.Timezone,
		Role:              string(enums.UserRoleCustomer),
		IsActive:          true,
		LastActiveAt:      nil,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if req.DateOfBirth != "" {
		dob, err := time.Parse("2006-01-02", req.DateOfBirth)
		if err == nil {
			user.DateOfBirth = &dob
		}
	}

	if req.FirstName != "" && req.LastName != "" {
		fullName := req.FirstName + " " + req.LastName
		user.FullName = &fullName
	}

	// Crear usuario en base de datos
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Crear cliente asociado (siempre existe un cliente para cada usuario)
	customer := &entities.Customer{
		PublicID:  uuid.New().String(),
		UserID:    &user.ID,
		FullName:  *user.FullName,
		Email:     user.Email,
		Phone:     user.Phone,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.customerRepo.Create(ctx, customer); err != nil {
		// Rollback: eliminar usuario creado
		_ = s.userRepo.Delete(ctx, user.ID)
		return nil, fmt.Errorf("failed to create customer profile: %w", err)
	}

	return user, nil
}

// Login autentica un usuario y crea una sesión
func (s *UserService) Login(ctx context.Context, req *request.LoginRequest) (*entities.Session, *entities.User, error) {
	// Validar request
	if req.Email == "" || req.Password == "" {
		return nil, nil, errors.New("email and password are required")
	}

	// Buscar usuario por email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, nil, errors.New("invalid credentials")
		}
		return nil, nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Verificar si el usuario está bloqueado
	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		return nil, nil, errors.New("account is locked")
	}

	// Verificar si el usuario está activo
	if !user.IsActive {
		return nil, nil, errors.New("account is inactive")
	}

	// Verificar contraseña
	if !s.hasher.VerifyPassword(user.PasswordHash, req.Password) {
		// Incrementar contador de intentos fallidos
		user.FailedLoginAttempts++
		user.UpdatedAt = time.Now()
		_ = s.userRepo.Update(ctx, user) // Ignoramos error, no crítico
		return nil, nil, errors.New("invalid credentials")
	}

	// Resetear contador de intentos fallidos
	user.FailedLoginAttempts = 0
	user.UpdatedAt = time.Now()
	if err := s.userRepo.Update(ctx, user); err != nil {
		// No es crítico, continuamos
	}

	// Actualizar último login (si el método existe en el repositorio)
	_ = s.userRepo.UpdateLastLogin(ctx, user.ID, "") // Ignoramos error

	// Crear sesión
	session := &entities.Session{
		SessionID:        uuid.New().String(),
		UserID:           user.ID,
		RefreshTokenHash: uuid.New().String(),
		IsValid:          true,
		ExpiresAt:        time.Now().Add(7 * 24 * time.Hour), // 7 días
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, user, nil
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
func (s *UserService) UpdateProfile(ctx context.Context, userID int64, req *request.UpdateUserRequest) (*entities.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Actualizar campos
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
func (s *UserService) ChangePassword(ctx context.Context, userID int64, req *request.ChangePasswordRequest) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return errors.New("user not found")
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Verificar contraseña actual
	if !s.hasher.VerifyPassword(user.PasswordHash, req.CurrentPassword) {
		return errors.New("current password is incorrect")
	}

	// Hashear nueva contraseña
	newHash, err := s.hasher.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	// Actualizar contraseña
	user.PasswordHash = newHash
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// Logout cierra una sesión específica
func (s *UserService) Logout(ctx context.Context, sessionID string) error {
	if err := s.sessionRepo.Invalidate(ctx, sessionID); err != nil {
		return fmt.Errorf("failed to logout: %w", err)
	}
	return nil
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

	// Marcar usuario como inactivo
	user.IsActive = false
	user.UpdatedAt = time.Now()

	// Invalidar todas las sesiones
	_ = s.sessionRepo.InvalidateAllForUser(ctx, userID)

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to delete account: %w", err)
	}

	return nil
}

// ============================================================================
// FUNCIONES HELPER PRIVADAS
// ============================================================================

// validateCreateUserRequest valida los datos de registro
func (s *UserService) validateCreateUserRequest(req *request.CreateUserRequest) error {
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
