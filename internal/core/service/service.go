package service

import (
	"context"
	"errors"
	"github.com/go-related/fileservice/internal/core/domain"
	cerror "github.com/go-related/fileservice/internal/core/errors"
	"github.com/go-related/fileservice/internal/core/ports"
	"github.com/sirupsen/logrus"
	"sync"
)

type PortService struct {
	mx   sync.Mutex
	repo ports.Repository
}

func NewPortService(repo ports.Repository) *PortService {
	return &PortService{sync.Mutex{}, repo}
}

func (svr *PortService) AddOrUpdatePorts(ctx context.Context, ports []domain.Port) ([]domain.Port, error) {
	if len(ports) == 0 {
		err := cerror.InvalidPortsInputs
		logrus.WithError(err).Error("invalid input")
		return nil, err
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	err := svr.StartTransaction(ctx)
	if err != nil {
		logrus.WithError(err).Error("error starting transaction")
		return nil, err
	}
	svr.mx.Lock()
	defer svr.mx.Unlock()
	var result []domain.Port
	for _, port := range ports {
		select {
		case <-ctx.Done():
			return nil, ctx.Err() //cancelled
		default:
		}
		insertedPort, err := svr.repo.AddOrUpdatePort(ctx, port)
		if err != nil {
			logrus.WithError(err).WithField("port_id", port.Id).Error("failed to save port")
			return nil, err
		}
		result = append(result, insertedPort)
	}
	return result, nil
}

func (svr *PortService) StartTransaction(ctx context.Context) error {
	if svr.repo.DoesTransactionExists() {
		return nil
	}
	svr.mx.Lock()
	defer svr.mx.Unlock()
	err := svr.repo.StartTransaction(ctx, true)
	if err != nil {
		logrus.WithError(err).Error("error starting transaction")
	}
	return err
}

func (svr *PortService) CommitTransaction(ctx context.Context) error {
	if !svr.repo.DoesTransactionExists() {
		err := errors.New("transaction not initiated, can't commit transaction")
		logrus.WithError(err)
		return err
	}
	svr.mx.Lock()
	defer svr.mx.Unlock()
	err := svr.repo.CommitTransaction(ctx)
	if err != nil {
		logrus.WithError(err).Error("error comitting transaction")
		return err
	}
	return nil
}

func (svr *PortService) AbortTransaction() error {
	svr.mx.Lock()
	defer svr.mx.Unlock()
	if svr.repo.DoesTransactionExists() {
		svr.repo.AbortTransaction()
	}
	return nil
}
