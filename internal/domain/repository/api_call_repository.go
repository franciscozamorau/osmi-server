// internal/domain/repository/api_call_repository.go
package repository

import (
	"context"

	apicall "github.com/franciscozamorau/osmi-server/internal/api/dto/api_call"
	commondto "github.com/franciscozamorau/osmi-server/internal/api/dto/common"
	"github.com/franciscozamorau/osmi-server/internal/domain/entities"
)

// APICallRepository define operaciones para llamadas API de integración
type APICallRepository interface {
	// Registro
	LogAPICall(ctx context.Context, call *entities.ApiCall) error

	// Búsquedas
	List(ctx context.Context, filter apicall.APICallFilter, pagination commondto.Pagination) ([]*entities.ApiCall, int64, error)
	FindByProvider(ctx context.Context, provider string, pagination commondto.Pagination) ([]*entities.ApiCall, int64, error)
	FindByEndpoint(ctx context.Context, endpoint string, pagination commondto.Pagination) ([]*entities.ApiCall, int64, error)
	FindByStatus(ctx context.Context, statusCode int, pagination commondto.Pagination) ([]*entities.ApiCall, int64, error)
	FindByUser(ctx context.Context, userID int64, pagination commondto.Pagination) ([]*entities.ApiCall, int64, error)
	FindFailedCalls(ctx context.Context, hours int) ([]*entities.ApiCall, error)
	FindSlowCalls(ctx context.Context, thresholdMs int, pagination commondto.Pagination) ([]*entities.ApiCall, int64, error)

	// Consultas específicas
	GetLastCallForProvider(ctx context.Context, provider, endpoint string) (*entities.ApiCall, error)
	GetCallsInPeriod(ctx context.Context, provider, endpoint string, startDate, endDate string) ([]*entities.ApiCall, error)
	GetRetryStatistics(ctx context.Context, provider, endpoint string) (*apicall.RetryStats, error)

	// Limpieza
	CleanOldAPICalls(ctx context.Context, retentionDays int) (int64, error)

	// Estadísticas
	GetAPICallStats(ctx context.Context, filter apicall.APICallFilter) (*apicall.APICallStatsResponse, error)
	GetProviderStats(ctx context.Context, provider string) (*apicall.ProviderAPICallStats, error)
	GetEndpointStats(ctx context.Context, endpoint string) (*apicall.EndpointStats, error)
	GetSuccessRate(ctx context.Context, provider, endpoint string) (float64, error)
	GetAverageResponseTime(ctx context.Context, provider, endpoint string) (float64, error)
	GetErrorRate(ctx context.Context, provider, endpoint string) (float64, error)
	GetMostFrequentErrors(ctx context.Context, provider, endpoint string, limit int) ([]*apicall.ErrorFrequency, error)
	GetPeakUsageTimes(ctx context.Context, provider string) ([]*apicall.UsagePeak, error)
}
