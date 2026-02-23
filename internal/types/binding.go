package types

import (
	"github.com/Egot3/Zhao/bindings"
	"github.com/Egot3/Zhao/exchanges"
	"github.com/Egot3/Zhao/queues"
)

type Binding struct {
	QueueName  string                 `toml:"queue-name"`
	RoutingKey string                 `toml:"routing-key"`
	Enabled    bool                   `toml:"enabled"`
	Exchange   string                 `toml:"exchange-name"`
	NoWait     bool                   `toml:"no-wait"`
	Args       map[string]interface{} `toml:"arguments"`
}

func (b Binding) Canonical(q queues.QueueStruct, e exchanges.ExchangeStruct) bindings.BindingStruct {
	return bindings.BindingStruct{
		Queue:      q,
		Exchange:   e,
		RoutingKey: b.RoutingKey,
	}
}
