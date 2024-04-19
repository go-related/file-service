package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-related/fileservice/internal/core/domain"
	"github.com/hashicorp/go-memdb"
	"github.com/sirupsen/logrus"
)

const (
	tableName = "ports"
)

var (
	// TransactionAlreadyExists this is a simplified version of what i want to achieve
	TransactionAlreadyExists = errors.New("transaction already exists")
)

type PortInMemoryRepository struct {
	trn *memdb.Txn
	db  *memdb.MemDB
}

func (rp *PortInMemoryRepository) AddOrUpdatePort(ctx context.Context, port domain.Port) (*domain.Port, error) {
	if rp.trn == nil { // Here maybe we can use Guards to check
		err := fmt.Errorf("please strart a transaction before countinuing") // maybe this needs to be a custom error
		logrus.Error(err)
		return nil, err
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	currentData, err := rp.GetById(ctx, port.Id)
	if err != nil {
		logrus.WithField("id", port.Id).WithError(err).Error("error loading port from db")
		return nil, err
	}
	// check if we have any cancellation before continuing
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
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
	return currentData, err
}

func (rp *PortInMemoryRepository) GetById(ctx context.Context, Id string) (*domain.Port, error) {
	if rp.trn == nil { // use Guard
		err := fmt.Errorf("please strart a transaction before countinuing") // maybe this needs to be a custom error
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
		return nil, nil
	}
	result := raw.(domain.Port)
	return &result, nil
}

func (rp *PortInMemoryRepository) StartTransaction(ctx context.Context) error {
	// check if we have any cancellation before continuing
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if rp.trn == nil { //otherwise we are inside a transaction
		rp.trn = rp.db.Txn(true)
	} else {
		return TransactionAlreadyExists
	}
	return nil
}

// CommitTransaction i would have prefered to pass a context to the commit so i would be able to cancel or timeout this operation in case it takes a long time
func (rp *PortInMemoryRepository) CommitTransaction(ctx context.Context) error {
	if rp.trn == nil {
		return nil // TODO return proper error
	}
	// check if we have any cancellation before continuing
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	rp.trn.Commit()
	rp.trn = nil // we do this so we know we have to start a transaction to update data on db
	return nil
}

func (rp *PortInMemoryRepository) AbortTransaction() {
	if rp.trn == nil {
		return // TODO return proper error
	}
	rp.trn.Abort()
	rp.trn = nil // same as in the commit
}

func NewPortRepository() (*PortInMemoryRepository, error) {
	db, err := memdb.NewMemDB(creatDbSchema())
	if err != nil {
		logrus.WithError(err).Error("error setting up memdb")
		return nil, err
	}
	repo := PortInMemoryRepository{
		db: db,
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
