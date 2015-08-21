package main

import (
	"github.com/fatih/color"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	//"sync"
)

var referenceManager *Manager

func findall(directory string) []string {
	var ret []string
	files, _ := ioutil.ReadDir(directory)
	for _, f := range files {
		filename := directory + "/" + f.Name()
		if f.IsDir() {
			ret = append(ret, findall(filename)...)
			continue
		}
		if strings.HasSuffix(f.Name(), ".yaml") {
			ret = append(ret, filename)
		}
	}
	return ret
}

func main() {
	//var wg sync.WaitGroup

	var files []string
	var found []string
	var findErr error

	if len(os.Args) > 1 {
		for _, file := range os.Args[1:] {
			if strings.HasSuffix(file, ".yaml") {
				found, findErr = filepath.Glob(file)
				if findErr != nil {
					log.Fatalf("Error finding file %v (%v)", file, findErr)
				}
			} else { //directory
				found = findall(os.Args[1])
			}
			files = append(files, found...)
		}
	} else {
		files, findErr = filepath.Glob("tests/*.yaml")
		if findErr != nil {
			log.Fatalf("Error reading yaml files in current directory (%v)", findErr)
		}
	}

	totalTests := len(files)
	passedTests := 0
	for _, filename := range files {
		// create new set of references for each file
		referenceManager = NewManager()
		color.Cyan("Running file %v", filename)
		contents, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Fatalf("Error reading file %v (%v)\n", filename, err)
		}
		contents = []byte(referenceManager.ParseData(string(contents)))
		m := make(Config)
		err = yaml.Unmarshal(contents, &m)
		if err != nil {
			log.Fatalf("Error decoding yaml (%v)\n", err)
		}

		clients := GetConfigs(m)
		steps := ParseLayout(m["layout"].(string), clients)
		errors := make([]error, len(steps))
		for idx, step := range steps {
			step.Run()
			errors[idx] = step.Err()
			//wg.Add(1)
			//go func(idx int, s *Step) {
			//	s.Run()
			//	errors[idx] = s.Err()
			//	defer wg.Done()
			//}(idx, step)
		}
		//wg.Wait()
		hasError := false
		for _, e := range errors {
			if e != nil {
				color.Magenta("Error on chain: %v", e)
				hasError = true
			}
		}

		if !hasError {
			color.Green("Test [%v] passed!\n\n", m["name"].(string))
			passedTests += 1
		} else {
			color.Red("Test [%v] FAILED!\n\n", m["name"].(string))
		}

	}

	if passedTests == totalTests {
		color.Green("Passed %v of %v tests!", passedTests, totalTests)
	} else if passedTests > 0 {
		color.Yellow("Passed %v of %v tests!", passedTests, totalTests)
	} else {
		color.Red("Passed NOTHING of %v tests!", totalTests)
	}
}
