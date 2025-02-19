package rest

type RestConfig struct {
	Domain string `mapstructure:"domain"`
	Port   int    `mapstructure:"port"`
	TLS    bool   `mapstructure:"https"`
}
