package types

type ServiceStruct struct {
	Name         string     `toml:"name"`
	Port         string     `toml:"port"`
	RabbitMQPort string     `toml:"rabbitPort"`
	Endpoints    []Endpoint `toml:"endpoints"`
}
