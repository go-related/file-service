package grpc

import (
	"context"
	"github.com/go-related/fileservice/internal/core/domain"
	"github.com/go-related/fileservice/internal/core/ports"
	"github.com/go-related/fileservice/proto/pb"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type PortsClient struct {
	client           pb.PortServiceClient
	streamJsonParser ports.StreamJsonParser
}

func NewPortClient(port string) *PortsClient {
	conn, err := grpc.Dial(port, grpc.WithInsecure())
	if err != nil {
		logrus.Fatalf("can not connect with server %v", err)
	}

	// create stream
	client := pb.NewPortServiceClient(conn)

	return &PortsClient{
		client: client,
	}
}

func (cl *PortsClient) ReadJsonFile(ctx context.Context, filepath string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure cancellation is called when main function exits

	cn := make(chan domain.Port)
	go func() {
		for {
			select {
			case <-ctx.Done():
				logrus.Info("Worker: Received cancellation signal. Exiting.")
				return
			case m := <-cn:
				logrus.WithField("id", m.Id).Info("New data read")
			}
		}
	}()
	err := cl.streamJsonParser.ReadJsonFile(ctx, filepath, cn)
	if err != nil {
		logrus.WithError(err).Error("error reading from file")
	}
}
