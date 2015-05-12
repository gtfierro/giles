package main

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
)

// regular expression for finding Client headers in the config files
var clientMatcher = regexp.MustCompile(`^Client:[0-9]+$`)

// Wrapper type to simplify later code
type Config map[interface{}]interface{}

// All clients in the config files should implement this interface, regardless of the underlying protocol
type Client interface {
	// performs the Input method for this client
	Input() error
	// verifies that the expected Output was met
	Output() error
	// returns the Client identifier
	GetID() int64
}

// returns a new client with the given @id using the given Config
func GetClient(id int64, cfg Config) (Client, error) {
	var c Client
	iface, found := cfg["Interface"]
	if !found {
		return c, fmt.Errorf("Could not find mandatory field Interface in %v", cfg)
	}

	switch iface {
	case "HTTP":
		return NewHTTPClient(id, cfg)
	}

	return c, nil
}

// given a config file, return a map of the clients
func GetConfigs(cfg Config) map[int64]Client {
	var clients = make(map[int64]Client)

	// for each section in the configuration file
	for section, contents := range cfg {
		// see if it matches Client:<num>
		if clientMatcher.MatchString(section.(string)) {
			clientIdStr := strings.Split(section.(string), ":")[1]
			clientId, convErr := strconv.ParseInt(clientIdStr, 10, 64)
			if convErr != nil {
				log.Fatalf("Could not identify clientId in string %v (%v)\n", clientIdStr, convErr)
			}
			client, clientErr := GetClient(clientId, contents.(Config))
			if clientErr != nil {
				log.Fatalf("Could not create client for config %v (%v)\n", contents, clientErr)
			}
			clients[clientId] = client
		}
	}
	return clients
}

// Given a config file, how do we parse the layout line and then set up the Clients for execution in the desired order?
// A execution layout will be 1 or more serial executions performed in parallel. The simplest serial chain is
//	1:Input -> 1:Output
// Which means: run the Input for Client 1, then wait for the Output for Client 1.
// If we have 2 clients, we can have them run in parallel or in serial.
// Parallel:
//	1:Input -> 1:Output; 2:Input -> 2:Output
// Serial:
//	1:Input -> 1:Output -> 2:Input -> 2:Output
// Alt serial:
//	1:Input -> 2:Input -> 1:Output -> 2:Output
//
// Aside from some syntactic sugar to handle loops/repetitions, this is probably good enough for the regression tests
// we're going to be running. Parsing should be fairly straightforward, providing that the syntax doesn't change much.
// The question now is how to implement!
// The Do() and Check() methods of the client have the same signature, namely
//	func() error
// So we could wrap them up in a struct and then chain calls
// 	type Step struct {
//		This	func() error
//		Next	Step
//	}
//

type Step struct {
	this func() error
	err  error
	next *Step
}

func NewStep(f func() error) *Step {
	s := &Step{this: f, err: nil}
	log.Printf("s %v", s)
	return s
}

func (s *Step) Then(f func() error) *Step {
	s2 := NewStep(f)
	s.next = s2
	return s2
}

func (s *Step) Run() {
	s.err = s.this()
	if s.next != nil {
		log.Printf("run %v", s.next)
		s.next.Run()
	}
}

// Now to parse the layout

func ParseLayout(layout string, clients map[int64]Client) []*Step {
	log.Printf("layout %v\n", layout)
	parallelChunks := strings.Split(layout, ";")
	log.Printf("num chunks %v\n", len(parallelChunks))
	steps := make([]*Step, len(parallelChunks))
	for idx, chunk := range parallelChunks {
		clientStrings := strings.Split(chunk, "->")
		var localStep *Step
		var nextStep *Step
		for chunkIdx, cs := range clientStrings {
			cs = strings.TrimSpace(cs)
			_split := strings.Split(cs, ":")
			clientIdStr, clientMethod := _split[0], _split[1]
			clientId, convErr := strconv.ParseInt(clientIdStr, 10, 64)
			if convErr != nil {
				log.Fatalf("Could not identify clientId in string %v (%v)\n", clientIdStr, convErr)
			}
			if chunkIdx == 0 {
				localStep = NewStep(getMethod(clientMethod, clients[clientId]))
			} else if chunkIdx == 1 {
				nextStep = localStep.Then(getMethod(clientMethod, clients[clientId]))
			} else {
				nextStep = nextStep.Then(getMethod(clientMethod, clients[clientId]))
			}
		}
		steps[idx] = localStep
	}
	return steps
}

func getMethod(method string, c Client) func() error {
	switch method {
	case "Input":
		return c.Input
	case "Output":
		return c.Output
	default:
		return noop
	}
}

func noop() error {
	return nil
}
