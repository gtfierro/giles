package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// This is a basic HTTP client that fulfills the expected Client interface
type HTTPClient struct {
	id               int64
	req              *http.Request
	expectedCode     int
	expectedContents string
	expectedFormat   string
	response         *http.Response
}

func NewHTTPClient(id int64, c Config) (*HTTPClient, error) {
	// handle input
	h := &HTTPClient{id: id}
	// grab the Input section of the client configuration
	cfgInput, ok := c["Input"].(Config)
	if !ok {
		d, _ := yaml.Marshal(&c)
		return h, fmt.Errorf("Input section for client was invalid: %v\n", string(d))
	}
	// parse out what we need to construct the request
	method, foundMethod := cfgInput["Method"]
	uri, foundUri := cfgInput["URI"]
	format, foundFormat := cfgInput["Format"]
	if !foundMethod || !foundUri || !foundFormat {
		return h, fmt.Errorf("Input section did not contain Method, URI and Format\n")
	}

	// encode the data to the right format
	var data []byte
	var encodeErr error
	switch format {
	case "JSON":
		toMarshal := make(map[string]interface{})
		datastring := referenceManager.ParseData(cfgInput["Data"].(string))
		dec := json.NewDecoder(bytes.NewReader([]byte(datastring)))
		dec.UseNumber()
		encodeErr = dec.Decode(&toMarshal)
		if encodeErr == nil {
			data, encodeErr = json.Marshal(toMarshal)
		}
	case "string":
		datastring := referenceManager.ParseData(cfgInput["Data"].(string))
		data = []byte(strings.TrimSpace(datastring))
	}
	if encodeErr != nil {
		return h, fmt.Errorf("Error encoding data %v as %v (%v)\n", cfgInput["Data"], format, encodeErr)
	}

	// construct the HTTP request
	h.req, _ = http.NewRequest(method.(string), uri.(string), bytes.NewBuffer(data))

	// handle the expected output configuration
	cfgOutput, ok := c["Output"].(Config)
	if !ok {
		d, _ := yaml.Marshal(&c)
		return h, fmt.Errorf("Output section for client was invalid: %v\n", string(d))
	}
	code, foundCode := cfgOutput["Code"]
	contents, foundContents := cfgOutput["Contents"]
	format, foundFormat = cfgOutput["Format"]
	if !foundCode || !foundContents || !foundFormat {
		d, _ := yaml.Marshal(&c)
		return h, fmt.Errorf("Output section for client was invalid: %v\n", string(d))
	}
	h.expectedCode = code.(int)
	h.expectedContents = referenceManager.ParseData(contents.(string))
	h.expectedFormat = format.(string)
	return h, nil
}

func (hc *HTTPClient) Input() error {
	var err error
	client := &http.Client{}
	hc.response, err = client.Do(hc.req)
	return err
}

func (hc *HTTPClient) Output() error {
	if hc.response == nil {
		return fmt.Errorf("Nil response")
	}
	if hc.response.StatusCode != hc.expectedCode {
		return fmt.Errorf("Status code was %v but expected %v\n", hc.response.StatusCode, hc.expectedCode)
	}
	var outputOK = false
	defer hc.response.Body.Close()
	contents, readErr := ioutil.ReadAll(hc.response.Body)
	if readErr != nil {
		return fmt.Errorf("Error when reading HTTP response body (%v)\n", readErr)
	}
	switch hc.expectedFormat {
	case "string":
		outputOK = checkString(string(contents), hc.expectedContents)
	case "JSON":
		outputOK = checkJSON(string(contents), hc.expectedContents)
	}
	if !outputOK {
		return fmt.Errorf("Contents were [%v] but expected [%v]\n", string(contents), hc.expectedContents)
	}
	return nil
}

func (hc *HTTPClient) GetID() int64 {
	return hc.id
}

// This is a basic HTTP client that handles streaming connections
type HTTPStreamClient struct {
	id               int64
	req              *http.Request
	expectedCode     []int
	expectedContents []string
	expectedFormat   []string
	response         *http.Response
	reader           *bufio.Reader
	outputIndex      int
}

func NewHTTPStreamClient(id int64, c Config) (*HTTPStreamClient, error) {
	// handle input
	h := &HTTPStreamClient{id: id, outputIndex: -1}
	// grab the Input section of the client configuration
	cfgInput, ok := c["Input"].(Config)
	if !ok {
		d, _ := yaml.Marshal(&c)
		return h, fmt.Errorf("Input section for client was invalid: %v\n", string(d))
	}
	// parse out what we need to construct the request
	method, foundMethod := cfgInput["Method"]
	uri, foundUri := cfgInput["URI"]
	format, foundFormat := cfgInput["Format"]
	if !foundMethod || !foundUri || !foundFormat {
		return h, fmt.Errorf("Input section did not contain Method, URI and Format\n")
	}

	// encode the data to the right format
	var data []byte
	var encodeErr error
	switch format {
	case "JSON":
		toMarshal := make(map[string]interface{})
		datastring := referenceManager.ParseData(cfgInput["Data"].(string))
		dec := json.NewDecoder(bytes.NewReader([]byte(datastring)))
		dec.UseNumber()
		encodeErr = dec.Decode(&toMarshal)
		if encodeErr == nil {
			data, encodeErr = json.Marshal(toMarshal)
		}
	case "string":
		datastring := referenceManager.ParseData(cfgInput["Data"].(string))
		data = []byte(strings.TrimSpace(datastring))
	}
	if encodeErr != nil {
		return h, fmt.Errorf("Error encoding data %v as %v (%v)\n", cfgInput["Data"], format, encodeErr)
	}

	// construct the HTTP request
	h.req, _ = http.NewRequest(method.(string), uri.(string), bytes.NewBuffer(data))

	// handle the expected output configuration
	cfgOutput, ok := c["Output"].(Config)
	if !ok {
		d, _ := yaml.Marshal(&c)
		return h, fmt.Errorf("Output section for client was invalid: %v\n", string(d))
	}
	for _, rawSection := range cfgOutput {
		section := rawSection.(Config)
		code, foundCode := section["Code"]
		contents, foundContents := section["Contents"]
		format, foundFormat = section["Format"]
		if !foundCode || !foundContents || !foundFormat {
			d, _ := yaml.Marshal(&c)
			return h, fmt.Errorf("Output section for client was invalid: %v\n", string(d))
		}
		h.expectedCode = append(h.expectedCode, code.(int))
		h.expectedContents = append(h.expectedContents, referenceManager.ParseData(contents.(string)))
		h.expectedFormat = append(h.expectedFormat, format.(string))
	}
	return h, nil
}

func (hc *HTTPStreamClient) Input() error {
	var err error
	client := &http.Client{}
	go func() {
		hc.response, err = client.Do(hc.req)
	}()
	return err
}

func (hc *HTTPStreamClient) Output() error {
	hc.outputIndex += 1
	if hc.response == nil {
		return fmt.Errorf("Nil response")
	}
	if hc.reader == nil {
		hc.reader = bufio.NewReader(hc.response.Body)
	}

	if hc.response.StatusCode != hc.expectedCode[hc.outputIndex] {
		return fmt.Errorf("Status code was %v but expected %v\n", hc.response.StatusCode, hc.expectedCode)
	}
	var outputOK = false
	var contents []byte
	var readErr error

	readBytes := make(chan []byte)
	go func() {
		contents, readErr = hc.reader.ReadBytes('\n')
		readBytes <- contents
	}()

	select {
	case <-readBytes:
		hc.reader.ReadBytes('\n') // discard second newline (sMAP delivers them in pairs)
	case <-time.After(2 * time.Second): // timeout
	}

	if readErr != nil {
		return fmt.Errorf("Error when reading HTTP response body (%v)\n", readErr)
	}
	switch hc.expectedFormat[hc.outputIndex] {
	case "string":
		outputOK = checkString(string(contents), hc.expectedContents[hc.outputIndex])
	case "JSON":
		outputOK = checkJSON(string(contents), hc.expectedContents[hc.outputIndex])
	}
	if !outputOK {
		return fmt.Errorf("Contents were [%v] but expected [%v]\n", string(contents), hc.expectedContents[hc.outputIndex])
	}
	return nil
}

func (hc *HTTPStreamClient) GetID() int64 {
	return hc.id
}
