package grpc

import (
	"github.com/go-related/fileservice/internal/core/domain"
	"github.com/go-related/fileservice/internal/core/ports"
	"github.com/go-related/fileservice/proto/pb"
	"github.com/sirupsen/logrus"
	"io"
)

type Server struct {
	pb.UnimplementedPortServiceServer
	service ports.Service
}

func NewServer(service ports.Service) *Server {
	return &Server{service: service}
}

func (srv *Server) CreateOrUpdatePorts(stream pb.PortService_CreateOrUpdatePortsServer) error {
	ctx := stream.Context()
	var failedMessagesNumber int64
	err := srv.service.StartTransaction(ctx)
	if err != nil {
		logrus.WithError(err).Error("failed to start transaction")
		return err
	}
	for {
		portMessage, err := stream.Recv()
		// handle error
		if err != nil {
			if err == io.EOF {
				err := srv.service.CommitTransaction(ctx)
				if err != nil {
					logrus.WithError(err).Error("failed to commit transaction")
					return err
				}
				return stream.SendAndClose(&pb.PortResponse{
					FailedItemsNumber: &failedMessagesNumber,
				})
			}
			logrus.WithError(err).Error("failed to load messages")
			abortError := srv.service.AbortTransaction()
			if abortError != nil {
				logrus.WithError(abortError)
			}
			return err
		}
		portData := convertPortRequestToDomain(portMessage)
		if portData != nil {
			// maybe the service should also take one element
			_, err := srv.service.AddOrUpdatePorts(ctx, []domain.Port{*portData})
			if err != nil {
				failedMessagesNumber++
				logrus.WithError(err).Error("failed to save the message")
			}
		} else {
			failedMessagesNumber++
		}
	}
}
