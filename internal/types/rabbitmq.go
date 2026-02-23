package types

type RabbitMQ struct {
	Queues    []Queue    `toml:"queues"`
	Exchanges []Exchange `toml:"exchanges"`
	Bindings  []Binding  `toml:"bindings"`
}
