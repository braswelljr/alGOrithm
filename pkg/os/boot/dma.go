package boot

// DMAResult compares moving data with programmed I/O and with DMA.
type DMAResult struct {
	Stages     []Stage `json:"stages"`
	PIOCPU     int     `json:"pioCpu"`     // CPU ticks spent copying with programmed I/O
	DMACPU     int     `json:"dmaCpu"`     // CPU ticks spent setting up DMA and handling its interrupt
	DMAFree    int     `json:"dmaFree"`    // CPU ticks free for other work during the DMA transfer
	Interrupts int     `json:"interrupts"` // DMA completion interrupts
}

// DMA moves bytes from a device to memory, wordSize bytes per bus transfer,
// two ways. With programmed I/O the CPU itself reads every word from the
// device and stores it, so it is busy for the whole transfer. With direct
// memory access the CPU programs the DMA controller (address, count) in a
// few ticks, the controller moves block bytes at a time on its own while the
// CPU runs other work, and raises one interrupt per block when it is done.
//
// Time: O(bytes / wordSize). Space: O(bytes / wordSize).
func DMA(bytes, wordSize, block int) (DMAResult, error) {
	if bytes < 1 || bytes > 1<<16 || (wordSize != 1 && wordSize != 2 && wordSize != 4 && wordSize != 8) || block < wordSize || block > bytes {
		return DMAResult{}, inputErr("1-65536 bytes, word size 1, 2, 4 or 8, block between the word size and the total")
	}
	words := (bytes + wordSize - 1) / wordSize
	var res DMAResult
	res.PIOCPU = 2 * words // a load and a store per word
	res.Stages = append(res.Stages, Stage{Lane: "CPU (programmed I/O)", What: "copying word by word", Start: 0, End: res.PIOCPU, State: "running"})
	t := 0
	setup := 3
	for done := 0; done < bytes; done += block {
		n := min(block, bytes-done)
		res.Stages = append(res.Stages, Stage{Lane: "CPU (DMA)", What: "program the controller", Start: t, End: t + setup, State: "running"})
		res.DMACPU += setup
		t += setup
		xfer := (n + wordSize - 1) / wordSize
		res.Stages = append(res.Stages, Stage{Lane: "DMA controller", What: "transfer a block", Start: t, End: t + xfer, State: "running"},
			Stage{Lane: "CPU (DMA)", What: "free for other work", Start: t, End: t + xfer, State: "idle"})
		res.DMAFree += xfer
		t += xfer
		res.Stages = append(res.Stages, Stage{Lane: "CPU (DMA)", What: "completion interrupt", Start: t, End: t + 1, State: "running"})
		res.DMACPU++
		res.Interrupts++
		t++
	}
	return res, nil
}
