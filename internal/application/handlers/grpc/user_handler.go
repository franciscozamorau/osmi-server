// internal/application/handlers/grpc/user_handler.go
package grpc

import (
	"context"
	"log"
	"time"

	osmi "github.com/franciscozamorau/osmi-protobuf/gen/pb"
	userdto "github.com/franciscozamorau/osmi-server/internal/api/dto/user"
	"github.com/franciscozamorau/osmi-server/internal/api/helpers"
	"github.com/franciscozamorau/osmi-server/internal/application/services"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserHandler struct {
	osmi.UnimplementedOsmiServiceServer
	userService *services.UserService
	jwtSecret   []byte
}

func NewUserHandler(userService *services.UserService, jwtSecret string) *UserHandler {
	return &UserHandler{
		userService: userService,
		jwtSecret:   []byte(jwtSecret),
	}
}

// CreateUser maneja la creación de un nuevo usuario
func (h *UserHandler) CreateUser(ctx context.Context, req *osmi.CreateUserRequest) (*osmi.UserResponse, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}
	if req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}
	if len(req.Password) < 6 {
		return nil, status.Error(codes.InvalidArgument, "password must be at least 6 characters")
	}

	createReq := &userdto.CreateUserRequest{
		Username: req.Name,
		Email:    req.Email,
		Password: req.Password,
		Role:     req.Role,
	}
	if createReq.Role == "" {
		createReq.Role = "customer"
	}

	user, err := h.userService.Register(ctx, createReq)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	roleName := "customer"
	if user.IsSuperuser {
		roleName = "admin"
	} else if user.IsStaff {
		roleName = "staff"
	}

	return &osmi.UserResponse{
		UserId:    user.PublicID,
		Status:    "active",
		Name:      helpers.SafeStringPtr(user.Username),
		Email:     user.Email,
		Role:      roleName,
		CreatedAt: timestamppb.New(user.CreatedAt),
	}, nil
}

// GetUser obtiene un usuario por ID
func (h *UserHandler) GetUser(ctx context.Context, req *osmi.GetUserRequest) (*osmi.UserResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	user, err := h.userService.GetUserByPublicID(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	roleName := "customer"
	if user.IsSuperuser {
		roleName = "admin"
	} else if user.IsStaff {
		roleName = "staff"
	}

	return &osmi.UserResponse{
		UserId:    user.PublicID,
		Status:    "active",
		Name:      helpers.SafeStringPtr(user.Username),
		Email:     user.Email,
		Role:      roleName,
		CreatedAt: timestamppb.New(user.CreatedAt),
	}, nil
}

// UpdateUser actualiza la información de un usuario
func (h *UserHandler) UpdateUser(ctx context.Context, req *osmi.UpdateUserRequest) (*osmi.UserResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	log.Printf("📝 UpdateUser: userId=%s, name=%v", req.UserId, req.Name)

	updateReq := &userdto.UpdateUserRequest{
		FirstName:         req.Name, // ✅ name → first_name
		Phone:             req.Phone,
		AvatarURL:         req.AvatarUrl,
		PreferredLanguage: req.PreferredLanguage,
		PreferredCurrency: req.PreferredCurrency,
		Timezone:          req.Timezone,
	}

	user, err := h.userService.UpdateUser(ctx, req.UserId, updateReq)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	roleName := "customer"
	if user.IsSuperuser {
		roleName = "admin"
	} else if user.IsStaff {
		roleName = "staff"
	}

	return &osmi.UserResponse{
		UserId:    user.PublicID,
		Status:    "active",
		Name:      helpers.SafeStringPtr(user.Username),
		Email:     user.Email,
		Role:      roleName,
		CreatedAt: timestamppb.New(user.CreatedAt),
	}, nil
}

// DeleteUser elimina (desactiva) un usuario
func (h *UserHandler) DeleteUser(ctx context.Context, req *osmi.DeleteUserRequest) (*osmi.Empty, error) {
	return nil, status.Error(codes.Unimplemented, "DeleteUser not implemented")
}

// ============================================================================
// LOGIN CON JWT
// ============================================================================

// Login autentica a un usuario y devuelve JWT
func (h *UserHandler) Login(ctx context.Context, req *osmi.LoginRequest) (*osmi.LoginResponse, error) {

	log.Printf("🔐 Login handler llamado con email: %s", req.Email)

	if req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}
	if req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	user, err := h.userService.Authenticate(ctx, req.Email, req.Password)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	claims := jwt.MapClaims{
		"user_id": user.PublicID,
		"email":   user.Email,
		"role":    user.Role,
		"exp":     expiresAt.Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(h.jwtSecret)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate token")
	}

	name := ""
	if user.Username != nil {
		name = *user.Username
	}

	return &osmi.LoginResponse{
		Token:     tokenString,
		ExpiresAt: timestamppb.New(expiresAt),
		User: &osmi.UserResponse{
			UserId:    user.PublicID,
			Status:    "active",
			Name:      name,
			Email:     user.Email,
			Role:      user.Role,
			CreatedAt: timestamppb.New(user.CreatedAt),
		},
	}, nil
}

// Logout cierra la sesión de un usuario
func (h *UserHandler) Logout(ctx context.Context, req *osmi.LogoutRequest) (*osmi.Empty, error) {
	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	err := h.userService.Logout(ctx, req.Token)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	log.Printf("✅ Logout exitoso")
	return &osmi.Empty{}, nil
}

// RefreshToken renueva el token de acceso
func (h *UserHandler) RefreshToken(ctx context.Context, req *osmi.RefreshTokenRequest) (*osmi.RefreshTokenResponse, error) {
	if req.RefreshToken == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh_token is required")
	}

	newToken, expiresAt, err := h.userService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	return &osmi.RefreshTokenResponse{
		Token:     newToken,
		ExpiresAt: timestamppb.New(expiresAt),
	}, nil
}

// ============================================================================
// FUNCIONES DE CONTEXTO
// ============================================================================

func (h *UserHandler) extractUserIDFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "metadata not found")
	}

	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return "", status.Error(codes.Unauthenticated, "authorization token not found")
	}

	tokenString := authHeaders[0]
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, status.Error(codes.Unauthenticated, "unexpected signing method")
		}
		return h.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return "", status.Error(codes.Unauthenticated, "invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "invalid token claims")
	}

	// El user_id ya es string, no necesita conversión adicional
	userID, ok := claims["user_id"].(string)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "user_id not found in token")
	}

	return userID, nil
}

func (h *UserHandler) extractSessionIDFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "metadata not found")
	}

	sessionHeaders := md.Get("x-session-id")
	if len(sessionHeaders) == 0 {
		return "", status.Error(codes.Unauthenticated, "session ID not found")
	}

	return sessionHeaders[0], nil
}

// ListUsers lista todos los usuarios
func (h *UserHandler) ListUsers(ctx context.Context, req *osmi.ListUsersRequest) (*osmi.UserListResponse, error) {
	page := int(req.Page)
	pageSize := int(req.PageSize)

	users, total, err := h.userService.ListUsers(ctx, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbUsers := make([]*osmi.UserResponse, 0, len(users))
	for _, user := range users {
		roleName := "customer"
		if user.IsSuperuser {
			roleName = "admin"
		} else if user.IsStaff {
			roleName = "staff"
		}

		pbUsers = append(pbUsers, &osmi.UserResponse{
			UserId:    user.PublicID,
			Status:    "active",
			Name:      helpers.SafeStringPtr(user.Username),
			Email:     user.Email,
			Role:      roleName,
			CreatedAt: timestamppb.New(user.CreatedAt),
		})
	}

	totalPages := int32(0)
	if pageSize > 0 {
		totalPages = int32((int(total) + pageSize - 1) / pageSize)
	}

	return &osmi.UserListResponse{
		Users:      pbUsers,
		TotalCount: int32(total),
		Page:       int32(page),
		PageSize:   int32(pageSize),
		TotalPages: totalPages,
	}, nil
}
