package main

import (
	"encoding/json"
	simplejson "github.com/bitly/go-simplejson"
	"log"
)

type SmapReading struct {
	Readings [][]uint64
	UUID     string
}

func processJSON(bytes *[]byte) [](*SmapReading) {
	var reading map[string]*json.RawMessage
	var dest [](*SmapReading)
	err := json.Unmarshal(*bytes, &reading)
	if err != nil {
		log.Panic(err)
		return nil
	}

	for _, v := range reading {
		if v == nil {
			continue
		}
		var sr SmapReading
		err = json.Unmarshal(*v, &sr)
		if err != nil {
			log.Panic(err)
			return nil
		}
		dest = append(dest, &sr)
	}
	return dest
}

func handleJSON(bytes *[]byte) {
	/*
	 * we receive a bunch of top-level keys that we don't know, so we unmarshal them into a
	 * map, and then parse each of the internal objects individually
	 */
	var message map[string]*json.RawMessage
	err := json.Unmarshal(*bytes, &message)
	if err != nil {
		log.Panic(err)
	}

	for path, reading := range message {

		println("path", path)

		js, err := simplejson.NewJson([]byte(*reading))
		if err != nil {
			log.Println(err)
		}

		// get uuid
		println(js.Get("uuid").MustString("default"))

		// get properties
		properties := js.Get("Properties").MustMap()
		for k, v := range properties {
			log.Println(k, ":", v.(string))
		}

		//get metadata
		metadata := js.Get("Metadata").MustMap()
		for k, v := range metadata {
			log.Println(k, ":", v.(string))
		}

		//TODO get actuator

	}
}
