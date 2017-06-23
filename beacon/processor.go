package beacon

import "sync"

// Processor is an interface used for background workers
type Processor interface {
	Start(*sync.WaitGroup)
}
