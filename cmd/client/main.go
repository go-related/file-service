package main

import (
	"bufio"
	"context"
	"fmt"
	igrpc "github.com/go-related/fileservice/internal/adapters/grpc"
	"github.com/go-related/fileservice/internal/adapters/parser"
	"github.com/go-related/fileservice/internal/core/ports"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
)

var (
	streamJsonParser ports.StreamJsonParser
)

var config clientConfig

func main() {
	initialize()
	runClient()
}

func runClient() {
	server := igrpc.NewPortClient(config.host, config.port, streamJsonParser)
	ctx, cancel := context.WithCancel(context.Background())

	fmt.Println("started to read from file")
	err := server.ReadJsonFile(ctx, config.filePath)
	if err != nil {
		logrus.WithError(err).Error("")
	}

	// Start a goroutine to listen for user input
	fmt.Println("Press 'c' to cancel or 'quit' to terminate")
	go func(c context.CancelFunc) {
		// Create a scanner to read from standard input
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			input := strings.ToLower(scanner.Text())
			if input == "c" {
				fmt.Println("Cancellation requested. Press 'quit' to terminate.")
				c()
			} else {
				fmt.Println("Unknown command. Press 'c' to cancel or 'quit' to terminate.")
			}
		}
	}(cancel)

	defer cancel()
	fmt.Println("Program terminated.")
}

func initialize() {
	// TODO take this from the yaml file
	config = clientConfig{
		addDelayAfterItem: true,
		filePath:          "/Users/user/projects/juligo/file-service/config/ports.json",
		port:              "50051",
		host:              "localhost",
	}
	streamJsonParser = parser.NewStreamJsonParser(config.addDelayAfterItem)
}

type clientConfig struct {
	addDelayAfterItem bool
	filePath          string
	host              string
	port              string
}
