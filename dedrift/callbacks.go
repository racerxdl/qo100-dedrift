package dedrift

type OnFFTCallback func(segFFT, fullFFT []float32)
type OnIQData func(data []complex64)
