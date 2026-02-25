package services

import (
	"context"
	"errors"
	"time"

	"github.com/franciscozamorau/osmi-server/internal/api/dto/request"
	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
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

func (s *UserService) Register(ctx context.Context, req *request.CreateUserRequest) (*entities.User, error) {
	// Validar que el email no exista
	existing, _ := s.userRepo.GetByEmail(ctx, req.Email)
	if existing != nil {
		return nil, errors.New("email already registered")
	}

	// Hashear contraseña
	passwordHash, err := s.hasher.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// Crear usuario
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
		IsActive:          true,
		LastActiveAt:      nil,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if req.DateOfBirth != "" {
		dob, _ := time.Parse("2006-01-02", req.DateOfBirth)
		user.DateOfBirth = &dob
	}

	if req.FirstName != "" && req.LastName != "" {
		fullName := req.FirstName + " " + req.LastName
		user.FullName = &fullName
	}

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	// Crear cliente asociado
	customer := &entities.Customer{
		PublicID:  uuid.New().String(),
		UserID:    &user.ID,
		FullName:  *user.FullName,
		Email:     user.Email,
		Phone:     user.Phone,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = s.customerRepo.Create(ctx, customer)
	if err != nil {
		// Rollback: eliminar usuario creado
		s.userRepo.Delete(ctx, user.ID)
		return nil, err
	}

	return user, nil
}

func (s *UserService) Login(ctx context.Context, req *request.LoginRequest) (*entities.Session, *entities.User, error) {
	// Buscar usuario por email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, nil, errors.New("invalid credentials")
	}

	// Verificar contraseña
	if !s.hasher.VerifyPassword(user.PasswordHash, req.Password) {
		// CORREGIDO: Incrementar contador de intentos fallidos (manualmente)
		// Como el método no existe, actualizamos el contador directamente
		user.FailedLoginAttempts++
		user.UpdatedAt = time.Now()
		s.userRepo.Update(ctx, user)
		return nil, nil, errors.New("invalid credentials")
	}

	// Verificar si el usuario está bloqueado
	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		return nil, nil, errors.New("account is locked")
	}

	// Verificar si el usuario está activo
	if !user.IsActive {
		return nil, nil, errors.New("account is inactive")
	}

	// CORREGIDO: Resetear contador de intentos fallidos (manualmente)
	user.FailedLoginAttempts = 0
	user.UpdatedAt = time.Now()
	s.userRepo.Update(ctx, user)

	// Actualizar último login (asumiendo que UpdateLastLogin existe)
	s.userRepo.UpdateLastLogin(ctx, user.ID, "")

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

	err = s.sessionRepo.Create(ctx, session)
	if err != nil {
		return nil, nil, err
	}

	return session, user, nil
}

func (s *UserService) GetProfile(ctx context.Context, userID int64) (*entities.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}

func (s *UserService) UpdateProfile(ctx context.Context, userID int64, req *request.UpdateUserRequest) (*entities.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
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
		dob, _ := time.Parse("2006-01-02", *req.DateOfBirth)
		user.DateOfBirth = &dob
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

	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) ChangePassword(ctx context.Context, userID int64, req *request.ChangePasswordRequest) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return errors.New("user not found")
	}

	// Verificar contraseña actual
	if !s.hasher.VerifyPassword(user.PasswordHash, req.CurrentPassword) {
		return errors.New("current password is incorrect")
	}

	// Hashear nueva contraseña
	newHash, err := s.hasher.HashPassword(req.NewPassword)
	if err != nil {
		return errors.New("failed to hash new password")
	}

	// Actualizar contraseña
	now := time.Now()
	user.PasswordHash = newHash
	// CORREGIDO: PasswordChangedAt eliminado (no existe en la entidad)
	user.UpdatedAt = now

	return s.userRepo.Update(ctx, user)
}

func (s *UserService) Logout(ctx context.Context, sessionID string) error {
	return s.sessionRepo.Invalidate(ctx, sessionID)
}

func (s *UserService) LogoutAll(ctx context.Context, userID int64) error {
	return s.sessionRepo.InvalidateAllForUser(ctx, userID)
}

func (s *UserService) DeleteAccount(ctx context.Context, userID int64) error {
	// Marcar usuario como inactivo en lugar de borrarlo
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return errors.New("user not found")
	}

	user.IsActive = false
	user.UpdatedAt = time.Now()

	// Invalidar todas las sesiones
	s.sessionRepo.InvalidateAllForUser(ctx, userID)

	return s.userRepo.Update(ctx, user)
}
