package parser

import (
	"context"
	"strings"
	"testing"
)

// convert this into a table test
func TestHappyPath(t *testing.T) {
	t.Run("ParsesCorrectlyMapIds", func(t *testing.T) {
		var jsonParser = NewStreamMapJsonParser(0)
		rawJson := `
		{
		  "AEAJM": {
			"name": "Ajman",
			"city": "Ajman",
			"country": "United Arab Emirates",
			"alias": [],
			"regions": [],
			"coordinates": [
			  55.5136433,
			  25.4052165
			],
			"province": "Ajman",
			"timezone": "Asia/Dubai",
			"unlocs": [
			  "AEAJM"
			],
			"code": "52000"
		  },
		  "AEAUH": {
			"name": "Abu Dhabi",
			"coordinates": [
			  54.37,
			  24.47
			],
			"city": "Abu Dhabi",
			"province": "Abu Z¸aby [Abu Dhabi]",
			"country": "United Arab Emirates",
			"alias": [],
			"regions": [],
			"timezone": "Asia/Dubai",
			"unlocs": [
			  "AEAUH"
			],
			"code": "52001"
		  }
		}
		`
		expectedIds := []string{"AEAJM", "AEAUH"}
		cn := jsonParser.Subscribe()

		//assertion
		go func() {
			counter := 0
			for newItem := range cn {
				if expectedIds[counter] != newItem.Id {
					t.Errorf("invalid id returned expected:'%s' got:'%s'", expectedIds[counter], newItem.Id)
				}
				counter++
				if counter > len(expectedIds) {
					break
				}
			}
		}()
		//execute

		err := jsonParser.readJsonFileFromReader(context.Background(), strings.NewReader(rawJson))
		if err != nil {
			t.Error(err)
		}
	})
}
