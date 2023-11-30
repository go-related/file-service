package ports

import (
	"context"
	"github.com/go-related/fileservice/internal/domain"
)

type SteamJsonParser interface {
	Subscribe() chan domain.Port
	ReadJsonFile(ctx context.Context, filePath string) error
}

type Repository interface {
	// AddOrUpdatePort now there are other ways to do a batch of them at the same time but for purposes of this, one by one should be ok
	AddOrUpdatePort(ctx context.Context, port domain.Port) error
	StartTransaction(ctx context.Context, canWrite bool) error
	DoesTransactionExists() bool
	CommitTransaction(ctx context.Context) error
	AbortTransaction()
}

type Service interface {
	AddOrUpdatePorts(ctx context.Context, port []domain.Port) error
	StartTransaction(ctx context.Context) error
	CommitTransaction(ctx context.Context) error
	AbortTransaction() error
}
