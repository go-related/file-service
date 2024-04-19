package ports

import (
	"context"
	"github.com/go-related/fileservice/internal/core/domain"
)

type StreamJsonParser interface {
	ReadJsonFile(ctx context.Context, filePath string, channel chan domain.Port) error
}

type Repository interface {
	// AddOrUpdatePort maybe add a option to add a list together
	AddOrUpdatePort(ctx context.Context, port domain.Port) (*domain.Port, error)
	StartTransaction(ctx context.Context) error
	CommitTransaction(ctx context.Context) error
	AbortTransaction()
}

type PortService interface {
	AddOrUpdatePorts(ctx context.Context, ports []domain.Port) ([]*domain.Port, error)
	StartTransaction(ctx context.Context) error
	CommitTransaction(ctx context.Context) error
	AbortTransaction() error
}
