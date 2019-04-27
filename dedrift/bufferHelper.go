package dedrift

func (w *Worker) checkAndResizeBuffers(length int) {
	if len(w.buffer0) < length {
		w.buffer0 = make([]complex64, length)
	}
	if len(w.buffer1) < length {
		w.buffer1 = make([]complex64, length)
	}
}

func swapAndTrimSlices(a *[]complex64, b *[]complex64, length int) {
	*a = (*a)[:length]
	*b = (*b)[:length]

	c := *b
	*b = *a
	*a = c
}
