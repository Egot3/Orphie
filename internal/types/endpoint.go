package types

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Endpoint struct {
	Path          string `toml:"path,required"`
	ParsedPath    string `toml:"-"`
	BenchmarkPath string `toml:"benchmarkPath,omitempty"` //<- completly normal code
	//HTMLOverPure  bool      `tom:"HTMLOverPure,omitempty"` <- second-lasting weakness
	Method        string    `toml:"method"`
	Timeout       string    `toml:"timeout"`
	Enabled       bool      `toml:"enabled"`
	ContentType   string    `toml:"contentType,omitempty"`
	LookAfter     []string  `toml:"lookAfter,omitempty"`
	UpdateLogPath string    `toml:"updateLogPath,omitempty"`
	LastUpdate    time.Time `toml:"lastUpdate"`
	//StandardPath  string    `toml:"standardPath,omitempty"` <-mental illness

	BenchmarkResponseHash [32]uint8 `toml:"benchmarkResponseHash,omitempty"`

	Params map[string]interface{} `toml:"params"`
}

func (e *Endpoint) ParsePathVariables() error {
	e.ParsedPath = e.Path

	separators := map[rune]bool{
		'/': true,
		'&': true,
		'=': true,
	}
	byPart := strings.FieldsFuncSeq(e.Path, func(r rune) bool {
		return separators[r]
	})

	for subdirectory := range byPart {
		if !strings.HasPrefix(subdirectory, ":") {
			continue
		}
		//log.Println(e)

		value, ok := e.Params[subdirectory[1:]].(int64)
		//log.Println("val: ", value)
		if !ok {
			return fmt.Errorf("Given subdirectory value is not found in .toml!")
		}

		e.ParsedPath = strings.ReplaceAll(e.Path, subdirectory, strconv.Itoa(int(value)))
	}
	return nil
}

func (e Endpoint) GetParsedVariables() []string {
	var vars []string

	separators := map[rune]bool{
		'/': true,
		'&': true,
		'=': true,
	}
	byPart := strings.FieldsFuncSeq(e.Path, func(r rune) bool {
		return separators[r]
	})
	//byPart := strings.SplitSeq(e.Path, "/")

	for subdirectory := range byPart {
		//log.Printf("subdir: %v\n", subdirectory)

		if !strings.HasPrefix(subdirectory, ":") {
			continue
		}
		vars = append(vars, subdirectory[1:])
	}
	//log.Println("vars: ", vars)
	return vars
}
