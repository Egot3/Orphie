package types

import (
	"fmt"
	"iter"
	"log"
)

type Config struct {
	Service ServiceStruct `toml:"service"`
}

func (cfg Config) GetSeq() iter.Seq[*Endpoint] {
	return func(yield func(*Endpoint) bool) {
		v := cfg.Service.Endpoints
		for i := range len(v) {
			if !yield(&v[i]) { //what do you mean consumer wants to stop consuming?
				return
			} //idiomatic moment
		}
	}
}

// Takes: config, endpoint's name(method|path), Pram name(like in struct) and new value
// Gives: Love and Trust, maybe a nil ptr
func SwitchParams[V []string | int](cfg *Config, epsName, paramName string, value V) error {
	for ep := range cfg.GetSeq() {

		err := ep.ParsePathVariables()
		if err != nil {
			return err
		}

		log.Println("if cont: ", epsName, "2:", (ep.Method + "|" + ep.ParsedPath))

		if epsName == (ep.Method + "|" + ep.ParsedPath) {
			ep.Params[paramName] = value
			return nil
		}
	}
	return fmt.Errorf("Param with name %v wasn't found", paramName)
}
