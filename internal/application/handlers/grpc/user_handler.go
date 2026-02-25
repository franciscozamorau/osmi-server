package grpc

import (
	"context"
	"strconv"

	osmi "github.com/franciscozamorau/osmi-protobuf/gen/pb"
	"github.com/franciscozamorau/osmi-server/internal/api/dto/request"
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
	// Convertir protobuf a DTO
	createReq := &request.CreateUserRequest{
		Username: req.Name,
		Email:    req.Email,
		Password: req.Password,
		Role:     req.Role,
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

// GetUser obtiene un usuario por su ID
func (h *UserHandler) GetUser(ctx context.Context, req *osmi.UserLookup) (*osmi.UserResponse, error) {
	// Validar que se proporcione un ID
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	// Convertir el ID de string a int64
	userID, err := strconv.ParseInt(req.UserId, 10, 64)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user_id format")
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

// ============================================================================
// MÉTODOS NO IMPLEMENTADOS EN EL PROTO ACTUAL
// ============================================================================

// Los siguientes métodos están comentados porque no existen en el proto actual.
// Si en el futuro se añaden al proto, se pueden descomentar y adaptar.

/*
func (h *UserHandler) UpdateUser(ctx context.Context, req *osmi.UpdateUserRequest) (*osmi.UserResponse, error) {
	return nil, status.Error(codes.Unimplemented, "UpdateUser not implemented in proto")
}

func (h *UserHandler) DeleteUser(ctx context.Context, req *osmi.Empty) (*osmi.Empty, error) {
	return nil, status.Error(codes.Unimplemented, "DeleteUser not implemented in proto")
}

func (h *UserHandler) Login(ctx context.Context, req *osmi.LoginRequest) (*osmi.LoginResponse, error) {
	return nil, status.Error(codes.Unimplemented, "Login not implemented in proto")
}

func (h *UserHandler) Logout(ctx context.Context, req *osmi.Empty) (*osmi.Empty, error) {
	return nil, status.Error(codes.Unimplemented, "Logout not implemented in proto")
}
*/

// ============================================================================
// FUNCIONES DE CONTEXTO (IMPLEMENTACIÓN PENDIENTE)
// ============================================================================

// extractUserIDFromContext extrae el userID del token JWT en el contexto
func (h *UserHandler) extractUserIDFromContext(ctx context.Context) (int64, error) {
	// Obtener el token del metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, status.Error(codes.Unauthenticated, "metadata not found")
	}

	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return 0, status.Error(codes.Unauthenticated, "authorization token not found")
	}

	// Quitar el prefijo "Bearer " si existe
	tokenString := authHeaders[0]
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}

	// Parsear y validar el token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, status.Error(codes.Unauthenticated, "unexpected signing method")
		}
		return h.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return 0, status.Error(codes.Unauthenticated, "invalid token")
	}

	// Extraer el userID del token
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

// extractSessionIDFromContext extrae el sessionID del contexto (implementación pendiente)
func (h *UserHandler) extractSessionIDFromContext(ctx context.Context) (string, error) {
	// Por ahora, retornar error
	return "", status.Error(codes.Unimplemented, "session extraction not implemented")
}
