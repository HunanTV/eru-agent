package defines

type DockerConfig struct {
	Endpoint string
	Ca       string
	Key      string
	Cert     string
	Health   int
}

type EruConfig struct {
	Endpoint string
}

type LenzConfig struct {
	Routes   string
	Forwards []string
	Stdout   bool
	Count    int
}

type MetricsConfig struct {
	Step      int64
	Timeout   int64
	Force     int64
	Transfers []string
}

type RedisConfig struct {
	Host string
	Port int
	Min  int
	Max  int
}

type VLanConfig struct {
	Physical []string
}

type APIConfig struct {
	Http   bool
	PubSub bool
	Addr   string
}

type LimitConfig struct {
	Memory uint64
}

type AgentConfig struct {
	HostName string `yaml:"hostname"`
	PidFile  string

	Docker  DockerConfig
	Eru     EruConfig
	Lenz    LenzConfig
	Metrics MetricsConfig
	VLan    VLanConfig
	Redis   RedisConfig
	API     APIConfig
	Limit   LimitConfig
}
