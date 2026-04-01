// internal/application/handlers/grpc/user_handler.go
package grpc

import (
	"context"
	"strconv"

	osmi "github.com/franciscozamorau/osmi-protobuf/gen/pb"
	userdto "github.com/franciscozamorau/osmi-server/internal/api/dto/user" // ← CAMBIADO
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

// ============================================================================
// MÉTODOS IMPLEMENTADOS
// ============================================================================

// CreateUser maneja la creación de un nuevo usuario
func (h *UserHandler) CreateUser(ctx context.Context, req *osmi.CreateUserRequest) (*osmi.UserResponse, error) {
	// Validar campos requeridos
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

	// Convertir protobuf a DTO
	createReq := &userdto.CreateUserRequest{ // ← CAMBIADO
		Username: req.Name,
		Email:    req.Email,
		Password: req.Password,
		Role:     req.Role,
	}
	if createReq.Role == "" {
		createReq.Role = "customer"
	}

	// Llamar al servicio
	user, err := h.userService.Register(ctx, createReq)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Convertir entidad a protobuf usando helpers
	return &osmi.UserResponse{
		UserId:    user.PublicID,
		Status:    "active",
		Name:      helpers.SafeStringPtr(user.Username),
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: timestamppb.New(user.CreatedAt),
	}, nil
}

// CORREGIDO: Ahora recibe GetUserRequest en lugar de UserLookup
func (h *UserHandler) GetUser(ctx context.Context, req *osmi.GetUserRequest) (*osmi.UserResponse, error) {
	// Validar que se proporcione un ID
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// Convertir el ID de string a int64
	userID, err := strconv.ParseInt(req.UserId, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format: must be a numeric ID")
	}

	// Llamar al servicio
	user, err := h.userService.GetProfile(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	// Convertir entidad a protobuf
	return &osmi.UserResponse{
		UserId:    user.PublicID,
		Status:    "active",
		Name:      helpers.SafeStringPtr(user.Username),
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: timestamppb.New(user.CreatedAt),
	}, nil
}

// UpdateUser actualiza la información de un usuario
func (h *UserHandler) UpdateUser(ctx context.Context, req *osmi.UpdateUserRequest) (*osmi.UserResponse, error) {
	return nil, status.Error(codes.Unimplemented, "UpdateUser not implemented")
}

// DeleteUser elimina (desactiva) un usuario
func (h *UserHandler) DeleteUser(ctx context.Context, req *osmi.DeleteUserRequest) (*osmi.Empty, error) {
	return nil, status.Error(codes.Unimplemented, "DeleteUser not implemented")
}

// Login autentica a un usuario
func (h *UserHandler) Login(ctx context.Context, req *osmi.LoginRequest) (*osmi.LoginResponse, error) {
	return nil, status.Error(codes.Unimplemented, "Login not implemented")
}

// Logout cierra la sesión de un usuario
func (h *UserHandler) Logout(ctx context.Context, req *osmi.LogoutRequest) (*osmi.Empty, error) {
	return nil, status.Error(codes.Unimplemented, "Logout not implemented")
}

// RefreshToken renueva el token de acceso
func (h *UserHandler) RefreshToken(ctx context.Context, req *osmi.RefreshTokenRequest) (*osmi.RefreshTokenResponse, error) {
	return nil, status.Error(codes.Unimplemented, "RefreshToken not implemented")
}

// ============================================================================
// FUNCIONES DE CONTEXTO PARA JWT
// ============================================================================

// extractUserIDFromContext extrae el userID del token JWT en el contexto
func (h *UserHandler) extractUserIDFromContext(ctx context.Context) (int64, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, status.Error(codes.Unauthenticated, "metadata not found")
	}

	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return 0, status.Error(codes.Unauthenticated, "authorization token not found")
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
		return 0, status.Error(codes.Unauthenticated, "invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, status.Error(codes.Unauthenticated, "invalid token claims")
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return 0, status.Error(codes.Unauthenticated, "user_id not found in token")
	}

	return int64(userIDFloat), nil
}

// extractSessionIDFromContext extrae el sessionID del contexto
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
