package httphandler

import (
	"fmt"
	"github.com/gtfierro/giles/archiver"
	"net/http"
	"os"
	"strings"
	"testing"
)

/** Archiver setup **/
var configfile = "httphandler_test.cfg"
var config = archiver.LoadConfig(configfile)
var a = archiver.NewArchiver(config)

func TestMain(m *testing.M) {
	go Handle(a, *config.HTTP.Port)
	os.Exit(m.Run())
}

func BenchmarkAddReading1(b *testing.B) {
	var jsonstringshort = `
	{
		"/fast/sensor0": {
			"Readings": [[9182731928374, 30]],
			"uuid": "b86df176-6b40-5d58-8f29-3b85f5cfbf1e"
		}
	}`
	api := "asdf"
	url := fmt.Sprintf("http://localhost:%v/add/%v", *config.HTTP.Port, api)
	for i := 0; i < b.N; i++ {
		readershort := strings.NewReader(jsonstringshort)
		http.Post(url, "application/json", readershort)
	}
}

func BenchmarkAddReadingFull1(b *testing.B) {
	var jsonstring = `
	{
		"/": {
			"Contents": [
				"fast"
			]
		},
		"/fast": {
			"Contents": [
				"sensor0"
			]
		},
		"/fast/sensor0": {
			"Properties": {
				"ReadingType": "long",
				"Timezone": "America/Los_Angeles",
				"UnitofMeasure": "V",
				"UnitofTime": "s"
			},
			"Metadata": {
				"Site": "Test Site",
				"Nested": {
					"key": "value",
					"other": "value"
				}
			},
			"Readings": [[9182731928374, 30]],
			"uuid": "b86df176-6b40-5d58-8f29-3b85f5cfbf1e"
		}
	}`
	api := "asdf"
	url := fmt.Sprintf("http://localhost:%v/add/%v", *config.HTTP.Port, api)
	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(jsonstring)
		http.Post(url, "application/json", reader)
	}
}
