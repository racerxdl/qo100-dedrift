package config

type SourceConfig struct {
	Address         string
	SampleRate      uint32
	CenterFrequency uint32
	Gain            uint8
}

type ServerConfig struct {
	RTLTCPAddress string
	HTTPAddress   string
	AllowControl  bool
}

type AGCConfig struct {
	AttackRate float32
	DecayRate  float32
	Reference  float32
	Gain       float32
	MaxGain    float32
}

type LoopConfig struct {
	Bandwidth float32
}

type TranslationConfig struct {
	TransitionWidth float64
	Gain            float64
}

type ProcessingConfig struct {
	BeaconOffset   float32
	WorkDecimation uint32
	AGC            AGCConfig
	CostasLoop     LoopConfig
	Translation    TranslationConfig
}

type ProgramConfig struct {
	Source     SourceConfig
	Server     ServerConfig
	Processing ProcessingConfig
}
