package grpc

import (
	"context"
	"fmt"
	"github.com/go-related/fileservice/internal/core/domain"
	"github.com/go-related/fileservice/internal/core/ports"
	"github.com/go-related/fileservice/proto/pb"
	"github.com/sirupsen/logrus"
	"io"
)

type PortsServer struct {
	pb.UnimplementedPortServiceServer
	portService ports.PortService
}

func NewPortServer(portService ports.PortService) *PortsServer {
	return &PortsServer{
		portService: portService,
	}
}

func (s *PortsServer) CreateOrUpdatePorts(stream pb.PortService_CreateOrUpdatePortsServer) error {
	ctx, cancel := context.WithCancel(stream.Context())
	defer func() {
		err := s.portService.AbortTransaction()
		if err != nil {
			logrus.WithError(err).Warn("aborting transaction.")
		}
		cancel()
	}()

	err := s.portService.StartTransaction(ctx)
	if err != nil {
		return err
	}
	var failedCount int64
	for {
		// check if we have any cancellation before continuing
		select {
		case <-ctx.Done():
			err = s.portService.AbortTransaction()
			return ctx.Err()
		default:
		}

		port, err := stream.Recv()
		if err == io.EOF {
			err := s.portService.CommitTransaction(ctx)
			if err != nil {
				return err
			}
			msg := ""
			if failedCount > 0 {
				msg = fmt.Sprintf("operation didn't complete successfully")
			} else {
				msg = "operation completed successfully"
			}
			return stream.SendAndClose(&pb.PortResponse{
				FailedItemsNumber: &failedCount,
				Message:           msg,
			})
		}
		if err != nil {
			return err
		}

		reqItems := convertPortRequestToDomain(port)
		if len(reqItems) > 0 {
			insertedItems, err := s.portService.AddOrUpdatePorts(ctx, reqItems)
			if err != nil {
				return s.CloseStreamWithError(stream, failedCount, "failed to store port data")
			}
			if len(insertedItems) != len(reqItems) {
				failedCount += int64(len(reqItems) - len(insertedItems))
			}
		}
	}
}

func (s *PortsServer) CloseStreamWithError(stream pb.PortService_CreateOrUpdatePortsServer, failedCount int64, msg string) error {
	err := s.portService.AbortTransaction()
	if err != nil {
		return err
	}
	return stream.SendAndClose(&pb.PortResponse{
		FailedItemsNumber: &failedCount,
		Message:           msg,
	})
}

func convertPortRequestToDomain(request *pb.PortRequest) []domain.Port {
	if request == nil {
		return nil
	}
	var result []domain.Port
	for key, item := range request.PortDetails {
		outputPort := domain.Port{
			Name:        item.Name,
			Id:          key,
			City:        item.City,
			Country:     item.Country,
			Alias:       item.Alias,
			Regions:     item.Regions,
			Coordinates: item.Coordinates,
			Province:    item.Province,
			Timezone:    item.Timezone,
			UNLOCs:      item.Unlocs,
			Code:        item.Code,
		}
		result = append(result, outputPort)
	}
	return result
}
