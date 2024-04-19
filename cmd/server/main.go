package main

import (
	"fmt"
	igrpc "github.com/go-related/fileservice/internal/adapters/grpc"
	"github.com/go-related/fileservice/internal/adapters/repository"
	"github.com/go-related/fileservice/internal/core/ports"
	"github.com/go-related/fileservice/internal/core/service"
	"github.com/go-related/fileservice/proto/pb"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
)

var (
	portService    ports.PortService
	portRepository ports.Repository
)

func main() {
	port := "50051" //TODO take this from configuration
	initializeDependencies()
	runServer(port)
}

func initializeDependencies() {
	repo, err := repository.NewPortRepository()
	if err != nil {
		logrus.WithError(err).Fatalf("couldn't initialize repository")
	}
	portRepository = repo
	portService = service.NewPortService(portRepository)
}

func runServer(port string) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		logrus.WithError(err).Fatalf("couldn't bind to the port")
	}
	server := grpc.NewServer()
	portServer := igrpc.NewPortServer(portService)
	pb.RegisterPortServiceServer(server, portServer)
	reflection.Register(server)
	logrus.Infof("server listening at %v", listener.Addr())
	if err := server.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
