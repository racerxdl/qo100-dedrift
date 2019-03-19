package rtltcp

type TunerType uint32

const (
	RtlsdrTunerUnknown TunerType = iota
	RtlsdrTunerE4000
	RtlsdrTunerFc0012
	RtlsdrTunerFc0013
	RtlsdrTunerFc2580
	RtlsdrTunerR820t
	RtlsdrTunerR828d
)
