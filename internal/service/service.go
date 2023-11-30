package service

import (
	"context"
	"errors"
	"github.com/go-related/fileservice/internal/domain"
	"github.com/go-related/fileservice/internal/ports"
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

/*

type Service interface {
	AddOrUpdatePorts(ctx context.Context,port []domain.Port) error
	StartTransaction(ctx context.Context) error
	CommitTransaction(ctx context.Context) error
	AbortTransaction() error
}
*/

func (svr *PortService) AddOrUpdatePorts(ctx context.Context, ports []domain.Port) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	err := svr.StartTransaction(ctx)
	if err != nil {
		logrus.WithError(err).Error("error starting transaction")
		return err
	}
	svr.mx.Lock()
	defer svr.mx.Lock()
	for _, port := range ports {
		select {
		case <-ctx.Done():
			return ctx.Err() //cancelled
		default:
		}
		err := svr.repo.AddOrUpdatePort(ctx, port)
		if err != nil {
			logrus.WithError(err).WithField("port_id", port.Id).Error("failed to save port")
			return err
		}
	}
	return nil
}

func (svr *PortService) StartTransaction(ctx context.Context) error {
	if svr.repo.DoesTransactionExists() {
		return nil
	}
	svr.mx.Lock()
	defer svr.mx.Lock()
	err := svr.repo.StartTransaction(ctx, true)
	if err != nil {
		logrus.WithError(err).Error("error starting transaction")
		return err
	}
	return nil
}

func (svr *PortService) CommitTransaction(ctx context.Context) error {
	if !svr.repo.DoesTransactionExists() {
		err := errors.New("transaction not initiated, can't commit transaction")
		logrus.WithError(err)
		return err
	}
	svr.mx.Lock()
	defer svr.mx.Lock()
	err := svr.repo.CommitTransaction(ctx)
	if err != nil {
		logrus.WithError(err).Error("error commiting transaction")
		return err
	}
	return nil
}

func (svr *PortService) AbortTransaction() error {

	if svr.repo.DoesTransactionExists() {
		svr.mx.Lock()
		defer svr.mx.Lock()
		svr.repo.AbortTransaction()
	}
	return nil
}
