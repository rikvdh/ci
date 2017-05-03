package builder

import (
	"github.com/rikvdh/ci/lib/config"
	"sync"
)

type jobCounter struct {
	mu         sync.RWMutex
	jobCounter uint
	jobLimit   uint
	eventCh    chan uint
}

func newJobCounter() *jobCounter {
	return &jobCounter{
		jobCounter: 0,
		jobLimit:   config.Get().ConcurrentBuilds,
		eventCh:    make(chan uint),
	}
}

func (j *jobCounter) Increment() {
	j.mu.Lock()
	j.jobCounter++
	j.mu.Unlock()
	j.publishEvent()
}

func (j *jobCounter) Decrement() {
	j.mu.Lock()
	j.jobCounter--
	j.mu.Unlock()
	j.publishEvent()
}

func (j *jobCounter) CanStartJob() bool {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return (j.jobCounter < j.jobLimit)
}

func (j *jobCounter) publishEvent() {
	j.mu.RLock()
	j.eventCh <- j.jobCounter
	j.mu.RUnlock()
}

func (j *jobCounter) GetEventChannel() <-chan uint {
	return j.eventCh
}
