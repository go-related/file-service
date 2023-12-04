package repository

import (
	"context"
	"github.com/go-related/fileservice/internal/core/domain"
	cerrors "github.com/go-related/fileservice/internal/core/errors"
	"github.com/hashicorp/go-memdb"
	"github.com/sirupsen/logrus"
	"sync"
)

const (
	tableName = "ports"
)

type PortInMemoryRepository struct {
	mx  sync.Mutex
	trn *memdb.Txn
	db  *memdb.MemDB
}

// AddOrUpdatePort  It might have been a better idea to return *domain.port
func (rp *PortInMemoryRepository) AddOrUpdatePort(ctx context.Context, port domain.Port) (domain.Port, error) {
	if rp.trn == nil { // Here maybe we can use Guards to check
		err := cerrors.TransactionRequeired
		logrus.Error(err)
		return domain.Port{}, err
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	rp.mx.Lock()
	defer rp.mx.Unlock()
	currentData, err := rp.getById(ctx, port.Id)
	if err != nil {
		logrus.WithField("id", port.Id).WithError(err).Error("error loading port from db")
		return domain.Port{}, err
	}
	// check if we have any cancellation before continuing
	select {
	case <-ctx.Done():
		return domain.Port{}, ctx.Err()
	default:
	}
	// here is the same code but i left intentionally to distinguish btw insert and update
	if currentData == nil {
		err = rp.trn.Insert(tableName, port)
		if err != nil {
			logrus.WithError(err).Error("error inserting data into table")
		}
	} else {
		err = rp.trn.Insert(tableName, port)
		if err != nil {
			logrus.WithError(err).Error("error updating data into table")
		}
	}
	return port, err
}

func (rp *PortInMemoryRepository) getById(ctx context.Context, Id string) (*domain.Port, error) {
	if rp.trn == nil { // use Guard
		err := cerrors.TransactionRequeired
		logrus.Error(err)
		return nil, err
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	raw, err := rp.trn.First(tableName, "id", Id)
	if err != nil {
		logrus.WithError(err).Error("error loading port from db")
		return nil, err
	}
	// check if we have any cancellation before continuing
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	//not found
	if raw == nil {
		return nil, cerrors.NotFound
	}
	result := raw.(domain.Port)
	return &result, nil
}

func (rp *PortInMemoryRepository) StartTransaction(ctx context.Context, canWrite bool) error {
	// check if we have any cancellation before continuing
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	rp.mx.Lock()
	defer rp.mx.Unlock()
	if rp.trn == nil { //otherwise we are inside a transaction
		rp.trn = rp.db.Txn(canWrite)
	} else {
		return cerrors.TransactionAlreadyExists
	}
	return nil
}

// CommitTransaction i would have prefered to pass a context to the commit so i would be able to cancel or timeout this operation in case it takes a long time
func (rp *PortInMemoryRepository) CommitTransaction(ctx context.Context) error {
	// check for cancel
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	rp.mx.Lock()
	defer rp.mx.Unlock()
	rp.trn.Commit()
	rp.trn = nil // we do this so we know we have to start a transaction to update data on db
	return nil
}

func (rp *PortInMemoryRepository) AbortTransaction() {
	if rp.trn != nil {
		rp.trn.Abort()
		rp.trn = nil // same as in the commit
	}
}

func (rp *PortInMemoryRepository) DoesTransactionExists() bool {
	return rp.trn != nil
}

func NewPortRepository() (*PortInMemoryRepository, error) {
	db, err := memdb.NewMemDB(creatDbSchema())
	if err != nil {
		logrus.WithError(err).Error("error setting up memdb")
		return nil, err
	}
	repo := PortInMemoryRepository{
		db: db,
		mx: sync.Mutex{},
	}
	return &repo, nil
}

func creatDbSchema() *memdb.DBSchema {
	return &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			tableName: {
				Name: tableName,
				Indexes: map[string]*memdb.IndexSchema{
					// Primary key index based on the 'Name' field, since its the identifier in our case
					"id": {
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.StringFieldIndex{Field: "Id"},
					},
				},
			},
		},
	}
}
