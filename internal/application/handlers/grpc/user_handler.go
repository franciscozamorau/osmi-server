// internal/application/handlers/grpc/user_handler.go
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
	createReq := &request.CreateUserRequest{
		Username: req.Name,
		Email:    req.Email,
		Password: req.Password,
		Role:     req.Role,
	}
	if createReq.Role == "" {
		createReq.Role = "customer" // Valor por defecto
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

// ============================================================================
// MÉTODOS NO IMPLEMENTADOS (PREPARADOS PARA EL FUTURO)
// ============================================================================

// UpdateUser actualiza la información de un usuario
// Nota: Este método no está en el proto actual. Cuando se agregue, se implementará aquí.
func (h *UserHandler) UpdateUser(ctx context.Context, req *osmi.UpdateUserRequest) (*osmi.UserResponse, error) {
	return nil, status.Error(codes.Unimplemented, "UpdateUser not implemented in proto")
}

// DeleteUser elimina (desactiva) un usuario
// Nota: Este método no está en el proto actual. Cuando se agregue, se implementará aquí.
func (h *UserHandler) DeleteUser(ctx context.Context, req *osmi.Empty) (*osmi.Empty, error) {
	return nil, status.Error(codes.Unimplemented, "DeleteUser not implemented in proto")
}

// Login autentica a un usuario y crea una sesión
// Nota: Este método no está en el proto actual. Cuando se agregue, se implementará aquí.
func (h *UserHandler) Login(ctx context.Context, req *osmi.LoginRequest) (*osmi.LoginResponse, error) {
	return nil, status.Error(codes.Unimplemented, "Login not implemented in proto")
}

// Logout cierra la sesión de un usuario
// Nota: Este método no está en el proto actual. Cuando se agregue, se implementará aquí.
func (h *UserHandler) Logout(ctx context.Context, req *osmi.Empty) (*osmi.Empty, error) {
	return nil, status.Error(codes.Unimplemented, "Logout not implemented in proto")
}

// ============================================================================
// FUNCIONES DE CONTEXTO PARA JWT
// ============================================================================

// extractUserIDFromContext extrae el userID del token JWT en el contexto
// Útil para interceptores y autenticación
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

// extractSessionIDFromContext extrae el sessionID del contexto
// Útil para interceptores y validación de sesiones
func (h *UserHandler) extractSessionIDFromContext(ctx context.Context) (string, error) {
	// Obtener el session_id del metadata
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
