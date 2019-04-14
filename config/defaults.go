package config

const (
	DefaultSourceAddress   = "127.0.0.1:1235"
	DefaultSampleRate      = 1800e3
	DefaultCenterFrequency = 740000000
	DefaultBeaconOffset    = 143e3
	DefaultWorkDecimation  = 32
)

const (
	DefaultRTLTCPAddress = ":1234"
	DefaultHTTPAddress   = ":8080"
	DefaultAllowControl  = true
)

const (
	DefaultTranslatorGain            = 64
	DefaultTranslatorTransitionWidth = 15e3
)

const (
	DefaultAGCAttackRate = 0.01
	DefaultAGCDecayRate  = 0.2
	DefaultAGCReference  = 1
	DefaultAGCGain       = 10
	DefaultAGCMaxGain    = 65535
)

const (
	DefaultCostasLoopBandwidth = 0.01
)

var DefaultConfig = ProgramConfig{
	Source: SourceConfig{
		Address:         DefaultSourceAddress,
		SampleRate:      DefaultSampleRate,
		CenterFrequency: DefaultCenterFrequency,
	},
	Server: ServerConfig{
		RTLTCPAddress: DefaultRTLTCPAddress,
		HTTPAddress:   DefaultHTTPAddress,
		AllowControl:  DefaultAllowControl,
	},
	Processing: ProcessingConfig{
		BeaconOffset:   DefaultBeaconOffset,
		WorkDecimation: DefaultWorkDecimation,
		AGC: AGCConfig{
			AttackRate: DefaultAGCAttackRate,
			DecayRate:  DefaultAGCDecayRate,
			Reference:  DefaultAGCReference,
			Gain:       DefaultAGCGain,
			MaxGain:    DefaultAGCMaxGain,
		},
		CostasLoop: LoopConfig{
			Bandwidth: DefaultCostasLoopBandwidth,
		},
		Translation: TranslationConfig{
			TransitionWidth: DefaultTranslatorTransitionWidth,
			Gain:            DefaultTranslatorGain,
		},
	},
}
