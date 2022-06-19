package cmd

type Metrics struct {
	totalMessages    uint64
	totalHeartbeats  uint64
	totalErrors      uint64
	bufferSize       int
	lastCommittedWal uint64
}

func NewMetrics() *Metrics {
	return &Metrics{
		totalMessages:    0,
		totalHeartbeats:  0,
		totalErrors:      0,
		bufferSize:       0,
		lastCommittedWal: 0,
	}
}
