package logger

import "sync"

type LogBuffer struct {
	mu   sync.RWMutex
	logs []string
	max  int
}

func NewBuffer(maxLines int) *LogBuffer {
	return &LogBuffer{
		logs: make([]string, 0, maxLines),
		max:  maxLines,
	}
}

func (b *LogBuffer) Add(line string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.logs) >= b.max {
		b.logs = b.logs[1:] // Shift out oldest log
	}
	b.logs = append(b.logs, line)
}

func (b *LogBuffer) Get() []string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// Return a copy to prevent data races
	cp := make([]string, len(b.logs))
	copy(cp, b.logs)
	return cp
}
