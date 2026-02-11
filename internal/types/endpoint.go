package types

import (
	"fmt"
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

	Params map[string]any `toml:",remain"`
}

func (e *Endpoint) parsePathVariables() error {
	byPart := strings.SplitSeq(e.Path, "/")
	for subdirectory := range byPart {
		if !strings.HasPrefix(subdirectory, ":") {
			continue
		}

		value, ok := e.Params[subdirectory[1:]].(string)
		if !ok {
			return fmt.Errorf("Given subdirectory value is not found in .toml!")
		}

		e.Path = strings.ReplaceAll(e.Path, subdirectory, value)
	}
	return nil
}
