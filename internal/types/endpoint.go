package types

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

type Endpoint struct {
	Path          string    `toml:"path"`
	ParsedPath    string    `toml:"-"`
	Method        string    `toml:"method"`
	Timeout       string    `toml:"timeout"`
	Enabled       bool      `toml:"enabled"`
	UpdateLogPath string    `toml:"updateLogPath"`
	LastUpdate    time.Time `toml:"lastUpdate"`

	Params map[string]interface{} `toml:"params"`
}

func (e *Endpoint) ParsePathVariables() error {
	e.ParsedPath = e.Path
	byPart := strings.SplitSeq(e.Path, "/")
	for subdirectory := range byPart {
		if !strings.HasPrefix(subdirectory, ":") {
			continue
		}
		log.Println(e)

		value, ok := e.Params[subdirectory[1:]].(int64)
		log.Println("val: ", value)
		if !ok {
			return fmt.Errorf("Given subdirectory value is not found in .toml!")
		}

		e.ParsedPath = strings.ReplaceAll(e.Path, subdirectory, strconv.Itoa(int(value)))
	}
	return nil
}

func (e Endpoint) GetParsedVariables() []string {
	var vars []string
	byPart := strings.SplitSeq(e.Path, "/")

	for subdirectory := range byPart {
		log.Printf("subdir: %v\n", subdirectory)

		if !strings.HasPrefix(subdirectory, ":") {
			continue
		}
		vars = append(vars, subdirectory[1:])
	}
	log.Println("vars: ", vars)
	return vars
}
