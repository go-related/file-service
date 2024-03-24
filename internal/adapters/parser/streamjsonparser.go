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

type StreamJsonParser struct {
	addDelayAfterItemRead bool
}

func NewStreamJsonParser(addDelayAfterItemRead bool) *StreamJsonParser {
	return &StreamJsonParser{addDelayAfterItemRead: addDelayAfterItemRead}
}

func (parser *StreamJsonParser) ReadJsonFile(ctx context.Context, filePath string, publishChannel chan domain.Port) error {
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
	return parser.readJsonFileFromReader(ctx, f, publishChannel)
}

func (parser *StreamJsonParser) readJsonFileFromReader(ctx context.Context, reader io.Reader, publishChannel chan domain.Port) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	decoder := jstream.NewDecoder(reader, 1).EmitKV() // extract JSON values at a depth level of 1
	for mv := range decoder.Stream() {
		//if we are cancelled or sm like that
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		keyValueData := mv.Value.(jstream.KV) // safe to do since we dit the EmitKV
		port := domain.Port{}
		mapData := keyValueData.Value.(map[string]interface{})
		err := convertToPort(mapData, &port)
		if err != nil {
			return err
		}
		// validate the extracted data and publish
		if len(port.Name) > 0 {
			port.Id = keyValueData.Key
			publishChannel <- port
		} else {
			return errors.New("couldn't parse data to the correct interface")
		}
		if parser.addDelayAfterItemRead {
			time.Sleep(5 * time.Second)
		}
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
