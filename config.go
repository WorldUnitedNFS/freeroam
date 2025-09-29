package freeroam

type UDPConfig struct {
	ListenAddress    string
	VisibilityRadius float64
	MaxVisiblePlayers int
}

type FMSConfig struct {
	ListenAddress  string
	AllowedOrigin  string
	UpdateInterval int
}

type Config struct {
	UDP UDPConfig
	FMS FMSConfig
}

func DefaultConfig() Config {
	return Config{
		UDP: UDPConfig{
			ListenAddress:     ":9999",
			VisibilityRadius:  300.0,
			MaxVisiblePlayers: 14,
		},
		FMS: FMSConfig{
			ListenAddress: "127.0.0.1:6996",
			AllowedOrigin: "127.0.0.1",
		},
	}
}
