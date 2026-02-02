package repository

import (
	"context"

	"github.com/franciscozamorau/osmi-server/internal/api/dto"
	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
)

// APICallRepository define operaciones para llamadas API de integración
type APICallRepository interface {
	// Registro
	LogAPICall(ctx context.Context, call *entities.APICall) error

	// Búsquedas
	List(ctx context.Context, filter dto.APICallFilter, pagination dto.Pagination) ([]*entities.APICall, int64, error)
	FindByProvider(ctx context.Context, provider string, pagination dto.Pagination) ([]*entities.APICall, int64, error)
	FindByEndpoint(ctx context.Context, endpoint string, pagination dto.Pagination) ([]*entities.APICall, int64, error)
	FindByStatus(ctx context.Context, statusCode int, pagination dto.Pagination) ([]*entities.APICall, int64, error)
	FindByUser(ctx context.Context, userID int64, pagination dto.Pagination) ([]*entities.APICall, int64, error)
	FindFailedCalls(ctx context.Context, hours int) ([]*entities.APICall, error)
	FindSlowCalls(ctx context.Context, thresholdMs int, pagination dto.Pagination) ([]*entities.APICall, int64, error)

	// Consultas específicas
	GetLastCallForProvider(ctx context.Context, provider, endpoint string) (*entities.APICall, error)
	GetCallsInPeriod(ctx context.Context, provider, endpoint string, startDate, endDate string) ([]*entities.APICall, error)
	GetRetryStatistics(ctx context.Context, provider, endpoint string) (*dto.RetryStats, error)

	// Limpieza
	CleanOldAPICalls(ctx context.Context, retentionDays int) (int64, error)

	// Estadísticas
	GetAPICallStats(ctx context.Context, filter dto.APICallFilter) (*dto.APICallStatsResponse, error)
	GetProviderStats(ctx context.Context, provider string) (*dto.ProviderAPICallStats, error)
	GetEndpointStats(ctx context.Context, endpoint string) (*dto.EndpointStats, error)
	GetSuccessRate(ctx context.Context, provider, endpoint string) (float64, error)
	GetAverageResponseTime(ctx context.Context, provider, endpoint string) (float64, error)
	GetErrorRate(ctx context.Context, provider, endpoint string) (float64, error)
	GetMostFrequentErrors(ctx context.Context, provider, endpoint string, limit int) ([]*dto.ErrorFrequency, error)
	GetPeakUsageTimes(ctx context.Context, provider string) ([]*dto.UsagePeak, error)
}
