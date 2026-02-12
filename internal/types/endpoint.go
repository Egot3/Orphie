package types

import (
	"fmt"
	"log"
	"strings"
	"time"
)

type Endpoint struct {
	Path          string    `toml:"path"`
	Method        string    `toml:"method"`
	Timeout       string    `toml:"timeout"`
	Enabled       bool      `toml:"enabled"`
	UpdateLogPath string    `toml:"updateLogPath"`
	LastUpdate    time.Time `toml:"lastUpdate"`

	Params map[string]interface{} `toml:",remain"`
}

func (e *Endpoint) ParsePathVariables() error {
	byPart := strings.SplitSeq(e.Path, "/")
	for subdirectory := range byPart {
		if !strings.HasPrefix(subdirectory, ":") {
			continue
		}
		log.Println(e)

		value, ok := e.Params[subdirectory[1:]]
		log.Println("val: ", value)
		if !ok {
			return fmt.Errorf("Given subdirectory value is not found in .toml!")
		}

		//e.Path = strings.ReplaceAll(e.Path, subdirectory, value)
	}
	return nil
}

func (e Endpoint) GetParsedVariables() []string {
	var vars []string
	byPart := strings.SplitSeq(e.Path, "/")
	for subdirectory := range byPart {
		if !strings.HasPrefix(subdirectory, ":") {
			continue
		}
		vars = append(vars, subdirectory[1:])
	}
	return vars
}
