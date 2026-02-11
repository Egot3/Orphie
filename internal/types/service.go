package types

type ServiceStruct struct {
	Name      string     `toml:"name"`
	Port      string     `toml:"port"`
	Endpoints []Endpoint `toml:"endpoints"`
}
