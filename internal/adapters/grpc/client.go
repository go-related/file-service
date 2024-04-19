package grpc

import (
	"context"
	"fmt"
	"github.com/go-related/fileservice/internal/core/domain"
	"github.com/go-related/fileservice/internal/core/ports"
	"github.com/go-related/fileservice/proto/pb"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"io"
)

type PortsClient struct {
	client           pb.PortServiceClient
	streamJsonParser ports.StreamJsonParser
}

func NewPortClient(host, port string, parser ports.StreamJsonParser) *PortsClient {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", host, port), grpc.WithInsecure())
	if err != nil {
		logrus.Fatalf("can not connect with server %v", err)
	}

	// create stream
	client := pb.NewPortServiceClient(conn)
	return &PortsClient{
		client:           client,
		streamJsonParser: parser,
	}
}

func (cl *PortsClient) ReadJsonFile(ctx context.Context, filepath string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
		fmt.Println("finished reading the file")
	}()

	cn := make(chan domain.Port)
	stream, err := cl.client.CreateOrUpdatePorts(ctx)
	if err != nil {
		return err
	}

	// set up channel reader and stream sender
	go func() {
		for {
			select {
			case <-ctx.Done():
				logrus.Info("Worker: Received cancellation signal. Exiting.")
				return
			case m := <-cn:
				logrus.WithField("id", m.Id).Info("New data read")
				req := &pb.PortRequest{
					PortDetails: map[string]*pb.PortDetails{
						m.Id: {
							Name:        m.Name,
							City:        m.City,
							Country:     m.Country,
							Alias:       m.Alias,
							Regions:     m.Regions,
							Coordinates: m.Coordinates,
							Province:    m.Province,
							Timezone:    m.Timezone,
							Unlocs:      m.UNLOCs,
							Code:        m.Code,
						},
					},
				}
				streamError := stream.Send(req)
				if streamError != nil {
					logrus.WithError(streamError).WithField("req", m).Error("error sending item to the stream")
				}
			}
		}
	}()

	// start reading from file and send to the channel
	err = cl.streamJsonParser.ReadJsonFile(ctx, filepath, cn)

	// setting up stream receiver
	go func() {
		for {
			select {
			case <-ctx.Done():
				closeErr := stream.CloseSend()
				if closeErr != nil {
					logrus.WithError(closeErr).Error("error closing the stream")
				}
			default:
			}
			var response pb.PortResponse
			err := stream.RecvMsg(response)
			if err == io.EOF {
				logrus.Info("connection closed from the server")
			}
			if err != nil {
				logrus.WithError(err).WithField("response", response).Error("error received from the server")
			}
			logrus.WithField("response", response).Info("response received from the server")
		}

	}()
	return err
}
