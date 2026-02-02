package grpchandlers

import (
	"context"
	"time"

	osmi "github.com/franciscozamorau/osmi-protobuf/gen/pb"
	"github.com/franciscozamorau/osmi-server/internal/api/dto"
	"github.com/franciscozamorau/osmi-server/internal/application/services"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserHandler struct {
	osmi.UnimplementedOsmiServiceServer
	userService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) CreateUser(ctx context.Context, req *osmi.UserRequest) (*osmi.UserResponse, error) {
	// Convertir protobuf a DTO
	createReq := &dto.CreateUserRequest{
		Email:             req.Email,
		Phone:             req.Phone,
		Username:          req.Username,
		Password:          req.Password,
		FirstName:         req.FirstName,
		LastName:          req.LastName,
		DateOfBirth:       req.DateOfBirth,
		PreferredLanguage: req.PreferredLanguage,
		PreferredCurrency: req.PreferredCurrency,
		Timezone:          req.Timezone,
	}

	// Llamar al servicio
	user, err := h.userService.Register(ctx, createReq)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Convertir entidad a protobuf
	return &osmi.UserResponse{
		Id:                user.PublicID,
		Email:             user.Email,
		Phone:             safeStringPtr(user.Phone),
		Username:          safeStringPtr(user.Username),
		FirstName:         safeStringPtr(user.FirstName),
		LastName:          safeStringPtr(user.LastName),
		FullName:          safeStringPtr(user.FullName),
		AvatarUrl:         safeStringPtr(user.AvatarURL),
		DateOfBirth:       safeTimeString(user.DateOfBirth),
		EmailVerified:     user.EmailVerified,
		PhoneVerified:     user.PhoneVerified,
		PreferredLanguage: user.PreferredLanguage,
		PreferredCurrency: user.PreferredCurrency,
		Timezone:          user.Timezone,
		MfaEnabled:        user.MFAEnabled,
		LastLoginAt:       safeTimeProto(user.LastLoginAt),
		IsActive:          user.IsActive,
		IsStaff:           user.IsStaff,
		IsSuperuser:       user.IsSuperuser,
		LastActiveAt:      timestamppb.New(user.LastActiveAt),
		CreatedAt:         timestamppb.New(user.CreatedAt),
		UpdatedAt:         timestamppb.New(user.UpdatedAt),
	}, nil
}

func (h *UserHandler) GetUser(ctx context.Context, req *osmi.UserLookup) (*osmi.UserResponse, error) {
	var userID string

	// Manejar diferentes formas de búsqueda
	switch lookup := req.Lookup.(type) {
	case *osmi.UserLookup_Id:
		userID = lookup.Id
	case *osmi.UserLookup_Email:
		// TODO: Implementar búsqueda por email
		return nil, status.Error(codes.Unimplemented, "search by email not implemented")
	default:
		return nil, status.Error(codes.InvalidArgument, "no valid lookup provided")
	}

	// Llamar al servicio
	user, err := h.userService.GetProfile(ctx, parseID(userID))
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	// Convertir entidad a protobuf
	return &osmi.UserResponse{
		Id:                user.PublicID,
		Email:             user.Email,
		Phone:             safeStringPtr(user.Phone),
		Username:          safeStringPtr(user.Username),
		FirstName:         safeStringPtr(user.FirstName),
		LastName:          safeStringPtr(user.LastName),
		FullName:          safeStringPtr(user.FullName),
		AvatarUrl:         safeStringPtr(user.AvatarURL),
		DateOfBirth:       safeTimeString(user.DateOfBirth),
		EmailVerified:     user.EmailVerified,
		PhoneVerified:     user.PhoneVerified,
		PreferredLanguage: user.PreferredLanguage,
		PreferredCurrency: user.PreferredCurrency,
		Timezone:          user.Timezone,
		MfaEnabled:        user.MFAEnabled,
		LastLoginAt:       safeTimeProto(user.LastLoginAt),
		IsActive:          user.IsActive,
		IsStaff:           user.IsStaff,
		IsSuperuser:       user.IsSuperuser,
		LastActiveAt:      timestamppb.New(user.LastActiveAt),
		CreatedAt:         timestamppb.New(user.CreatedAt),
		UpdatedAt:         timestamppb.New(user.UpdatedAt),
	}, nil
}

func (h *UserHandler) UpdateUser(ctx context.Context, req *osmi.UpdateUserRequest) (*osmi.UserResponse, error) {
	// Convertir protobuf a DTO
	updateReq := &dto.UpdateUserRequest{
		Phone:             req.Phone,
		FirstName:         req.FirstName,
		LastName:          req.LastName,
		AvatarURL:         req.AvatarUrl,
		DateOfBirth:       req.DateOfBirth,
		PreferredLanguage: req.PreferredLanguage,
		PreferredCurrency: req.PreferredCurrency,
		Timezone:          req.Timezone,
	}

	// Obtener userID del contexto (debería estar en los metadatos de autenticación)
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	// Llamar al servicio
	user, err := h.userService.UpdateProfile(ctx, userID, updateReq)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// Convertir entidad a protobuf
	return &osmi.UserResponse{
		Id:                user.PublicID,
		Email:             user.Email,
		Phone:             safeStringPtr(user.Phone),
		Username:          safeStringPtr(user.Username),
		FirstName:         safeStringPtr(user.FirstName),
		LastName:          safeStringPtr(user.LastName),
		FullName:          safeStringPtr(user.FullName),
		AvatarUrl:         safeStringPtr(user.AvatarURL),
		DateOfBirth:       safeTimeString(user.DateOfBirth),
		EmailVerified:     user.EmailVerified,
		PhoneVerified:     user.PhoneVerified,
		PreferredLanguage: user.PreferredLanguage,
		PreferredCurrency: user.PreferredCurrency,
		Timezone:          user.Timezone,
		MfaEnabled:        user.MFAEnabled,
		LastLoginAt:       safeTimeProto(user.LastLoginAt),
		IsActive:          user.IsActive,
		IsStaff:           user.IsStaff,
		IsSuperuser:       user.IsSuperuser,
		LastActiveAt:      timestamppb.New(user.LastActiveAt),
		CreatedAt:         timestamppb.New(user.CreatedAt),
		UpdatedAt:         timestamppb.New(user.UpdatedAt),
	}, nil
}

func (h *UserHandler) DeleteUser(ctx context.Context, req *osmi.Empty) (*osmi.Empty, error) {
	// Obtener userID del contexto
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	// Llamar al servicio
	err = h.userService.DeleteAccount(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &osmi.Empty{}, nil
}

func (h *UserHandler) Login(ctx context.Context, req *osmi.LoginRequest) (*osmi.LoginResponse, error) {
	// Convertir protobuf a DTO
	loginReq := &dto.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
		DeviceID: req.DeviceId,
	}

	// Llamar al servicio
	session, user, err := h.userService.Login(ctx, loginReq)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, err.Error())
	}

	// TODO: Generar tokens JWT
	accessToken := "generated_access_token"
	refreshToken := session.RefreshTokenHash

	return &osmi.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: &osmi.UserResponse{
			Id:                user.PublicID,
			Email:             user.Email,
			Phone:             safeStringPtr(user.Phone),
			Username:          safeStringPtr(user.Username),
			FirstName:         safeStringPtr(user.FirstName),
			LastName:          safeStringPtr(user.LastName),
			FullName:          safeStringPtr(user.FullName),
			EmailVerified:     user.EmailVerified,
			PhoneVerified:     user.PhoneVerified,
			PreferredLanguage: user.PreferredLanguage,
			PreferredCurrency: user.PreferredCurrency,
			Timezone:          user.Timezone,
			MfaEnabled:        user.MFAEnabled,
			IsActive:          user.IsActive,
			LastActiveAt:      timestamppb.New(user.LastActiveAt),
		},
		SessionId: session.SessionID,
		ExpiresAt: timestamppb.New(session.ExpiresAt),
	}, nil
}

func (h *UserHandler) Logout(ctx context.Context, req *osmi.Empty) (*osmi.Empty, error) {
	// Obtener sessionID del contexto
	sessionID, err := getSessionIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}

	// Llamar al servicio
	err = h.userService.Logout(ctx, sessionID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &osmi.Empty{}, nil
}

// Helper functions
func safeStringPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func safeTimeString(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

func safeTimeProto(t *time.Time) *timestamppb.Timestamp {
	if t == nil || t.IsZero() {
		return nil
	}
	return timestamppb.New(*t)
}

func parseID(id string) int64 {
	// TODO: Implementar parsing de ID
	return 0
}

func getUserIDFromContext(ctx context.Context) (int64, error) {
	// TODO: Extraer userID del contexto (de JWT)
	return 0, nil
}

func getSessionIDFromContext(ctx context.Context) (string, error) {
	// TODO: Extraer sessionID del contexto
	return "", nil
}
