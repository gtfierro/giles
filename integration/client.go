package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
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
		dec := json.NewDecoder(bytes.NewReader([]byte(cfgInput["Data"].(string))))
		dec.UseNumber()
		dec.Decode(&toMarshal)
		data, encodeErr = json.Marshal(toMarshal)
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
	h.expectedContents = contents.(string)
	switch format {
	case "string":
	case "JSON":
	}
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
	contents, err := ioutil.ReadAll(hc.response.Body)
	if err != nil {
		return err
	}
	if string(contents) != hc.expectedContents {
		return fmt.Errorf("Contents were [%v] but expected [%v]\n", string(contents), hc.expectedContents)
	}
	return nil
}

func (hc *HTTPClient) GetID() int64 {
	return hc.id
}
