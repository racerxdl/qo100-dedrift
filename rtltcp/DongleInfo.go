package rtltcp

type DongleInfo struct {
	Magic          [4]uint8
	TunerType      TunerType
	TunerGainCount uint32
}
