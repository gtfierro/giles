package main

import (
	"github.com/fatih/color"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
)

var referenceManager = NewManager()

func main() {
	var wg sync.WaitGroup

	var files []string
	var findErr error

	if len(os.Args) > 1 {
		for _, file := range os.Args[1:] {
			found, findErr := filepath.Glob(file)
			if findErr != nil {
				log.Fatalf("Error finding file %v (%v)", file, findErr)
			}
			files = append(files, found...)
		}
	} else {
		files, findErr = filepath.Glob("tests/*.yaml")
		if findErr != nil {
			log.Fatalf("Error reading yaml files in current directory (%v)", findErr)
		}
	}

	for _, filename := range files {
		color.Cyan("Running file %v", filename)
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
		errors := make([]error, len(steps))
		for idx, step := range steps {
			wg.Add(1)
			go func(idx int, s *Step) {
				s.Run()
				errors[idx] = s.Err()
				defer wg.Done()
			}(idx, step)
		}
		wg.Wait()
		hasError := false
		for _, e := range errors {
			if e != nil {
				color.Red("Error on chain: %v", e)
				hasError = true
			}
		}

		if !hasError {
			color.Green("Test [%v] passed!\n\n", m["name"].(string))
		}

	}
}
