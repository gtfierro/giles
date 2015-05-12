package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"sync"
)

func main() {
	var wg sync.WaitGroup

	filename := "test.yaml"
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Error reading file %v (%v)\n", filename, err)
	}
	m := make(Config)
	err = yaml.Unmarshal(contents, &m)
	if err != nil {
		log.Fatalf("Error decoding yaml (%v)\n", err)
	}

	clients := GetConfigs(m)
	steps := ParseLayout(m["layout"].(string), clients)
	for _, step := range steps {
		wg.Add(1)
		go func(s *Step) {
			s.Run()
			defer wg.Done()
		}(step)
	}
	wg.Wait()
}
