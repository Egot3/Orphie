package types

import "github.com/Egot3/Zhao/exchanges"

type Exchange struct {
	Name        string                 `toml:"name"`
	Enabled     bool                   `toml:"enabled"`
	Type        string                 `toml:"type"`
	Durable     bool                   `toml:"durable"`
	AutoDeleted bool                   `toml:"auto-deleted"`
	Internal    bool                   `toml:"internal"`
	NoWait      bool                   `toml:"no-wait"`
	Args        map[string]interface{} `toml:"arguments"`
}

func (e Exchange) Canonical() exchanges.ExchangeStruct {
	return exchanges.ExchangeStruct{
		Name:        e.Name,
		Type:        e.Type,
		Durable:     e.Durable,
		AutoDeleted: e.AutoDeleted,
		Internal:    e.Internal,
		NoWait:      e.NoWait,
		Args:        e.Args,
	}
}
