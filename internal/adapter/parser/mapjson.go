package parser

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/bcicen/jstream"
	"github.com/go-related/fileservice/internal/core/domain"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"time"
)

type StreamMapJsonParser struct {
	publishChannel chan domain.Port
	sleepTime      time.Duration
}

func NewStreamMapJsonParser(sleepAfterEachItem int64) *StreamMapJsonParser {
	srv := StreamMapJsonParser{
		publishChannel: make(chan domain.Port), //we can also make this a buffered channel but for now i am gonna leave it like this
	}
	if sleepAfterEachItem != 0 {
		sleepTime := sleepAfterEachItem * time.Second.Nanoseconds()
		srv.sleepTime = time.Duration(sleepTime)
	}
	return &srv
}

func (parser *StreamMapJsonParser) Subscribe() chan domain.Port {
	return parser.publishChannel
}

func (parser *StreamMapJsonParser) ReadJsonFile(ctx context.Context, filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		logrus.WithError(err).Error("error reading json file")
		return err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			logrus.WithError(err).Error("error closing reader to the file")
		}
	}(f)
	return parser.readJsonFileFromReader(ctx, f)
}

func (parser *StreamMapJsonParser) readJsonFileFromReader(ctx context.Context, reader io.Reader) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	decoder := jstream.NewDecoder(reader, 1).EmitKV() // extract JSON values at a depth level of 1
	for mv := range decoder.Stream() {
		select {
		case <-ctx.Done():
			return ctx.Err() //if we are cancelled or sm like that
		default:
		}
		// extract the data
		keyValueData := mv.Value.(jstream.KV) // safe to do since we dit the EmitKV
		port := domain.Port{}
		mapData := keyValueData.Value.(map[string]interface{}) //we know this si here we parse a struct
		err := convertToPort(mapData, &port)
		if err != nil {
			return err
		}
		// validate the extracted data and publish
		if len(port.Name) > 0 {
			port.Id = keyValueData.Key
			parser.publishChannel <- port
		} else {
			return errors.New("couldn't parse data to the correct interface")
		}
		time.Sleep(parser.sleepTime)
	}
	return nil
}

func convertToPort(data map[string]interface{}, target *domain.Port) error {
	// Convert the map to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	// Unmarshal the JSON into the struct
	if err := json.Unmarshal(jsonData, target); err != nil {
		return err
	}
	return nil
}
