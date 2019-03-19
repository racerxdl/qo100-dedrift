package rtltcp

type CommandType uint8

const (
	SetFrequency           CommandType = 0x01
	SetSampleRate                      = 0x02
	SetGainMode                        = 0x03
	SetGain                            = 0x04
	SetFrequencyCorrection             = 0x05
	SetIfStage                         = 0x06
	SetTestMode                        = 0x07
	SetAgcMode                         = 0x08
	SetDirectSampling                  = 0x09
	SetOffsetTuning                    = 0x0A
	SetRtlCrystal                      = 0x0B
	SetTunerCrystal                    = 0x0C
	SetTunerGainByIndex                = 0x0D
	SetTunerBandwidth                  = 0x0E
	SetBiasTee                         = 0x0F

	Invalid = 0xFF
)

var CommandTypeToName = map[CommandType]string{
	SetFrequency:           "SetFrequency",
	SetSampleRate:          "SetSampleRate",
	SetGainMode:            "SetGainMode",
	SetGain:                "SetGain",
	SetFrequencyCorrection: "SetFrequencyCorrection",
	SetIfStage:             "SetIfStage",
	SetTestMode:            "SetTestMode",
	SetAgcMode:             "SetAgcMode",
	SetDirectSampling:      "SetDirectSampling",
	SetOffsetTuning:        "SetOffsetTuning",
	SetRtlCrystal:          "SetRtlCrystal",
	SetTunerCrystal:        "SetTunerCrystal",
	SetTunerGainByIndex:    "SetTunerGainByIndex",
	SetTunerBandwidth:      "SetTunerBandwidth",
	SetBiasTee:             "SetBiasTee",
}

type Command struct {
	Type  CommandType
	Param [4]byte
}
