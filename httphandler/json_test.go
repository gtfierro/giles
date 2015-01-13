package httphandler

import (
	"strings"
	"testing"
)

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

var reader = strings.NewReader(jsonstring)

var jsonstringshort = `
{
    "/fast/sensor0": {
        "Readings": [[9182731928374, 30]],
        "uuid": "b86df176-6b40-5d58-8f29-3b85f5cfbf1e"
    }
}`
var readershort = strings.NewReader(jsonstringshort)

func BenchmarkDecodeJSONFull(b *testing.B) {
	for i := 0; i < b.N; i++ {
		handleJSON(reader)
	}
}

func BenchmarkDecodeJSONShort(b *testing.B) {
	for i := 0; i < b.N; i++ {
		handleJSON(readershort)
	}
}
