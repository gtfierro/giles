package httphandler

import (
	"fmt"
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

var jsonstringshort_nano = `
{
    "/fast/sensor0": {
        "Readings": [[9182731928374111111, 30]],
        "uuid": "b86df176-6b40-5d58-8f29-3b85f5cfbf1e"
    }
}`

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

func BenchmarkDecodeFFJSONFull(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ffhandleJSON(reader)
	}
}

func BenchmarkDecodeFFJSONShort(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ffhandleJSON(readershort)
	}
}

func TestDecodeShort(t *testing.T) {
	var readershort = strings.NewReader(jsonstringshort)
	res, err := handleJSON(readershort)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(res)
	for path, msg := range res {
		fmt.Println(path, msg)
	}

	readershort = strings.NewReader(jsonstringshort)
	res2, err := ffhandleJSON(readershort)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(res2)
	for path, msg := range res2 {
		fmt.Println(path, msg)
	}

}

func TestDecode(t *testing.T) {
	var reader = strings.NewReader(jsonstring)
	res, err := handleJSON(reader)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(res)
	for path, msg := range res {
		fmt.Println(path, msg)
	}
	fmt.Println()

	reader = strings.NewReader(jsonstring)
	res2, err := ffhandleJSON(reader)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(res2)
	for path, msg := range res2 {
		fmt.Println(path, msg)
	}

}

func TestDecodeNano(t *testing.T) {
	var reader = strings.NewReader(jsonstringshort_nano)
	res, err := handleJSON(reader)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(res)
	for path, msg := range res {
		fmt.Println(path, msg)
	}
	fmt.Println()

	reader = strings.NewReader(jsonstringshort_nano)
	res2, err := ffhandleJSON(reader)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(res2)
	for path, msg := range res2 {
		fmt.Println(path, msg)
	}

}
