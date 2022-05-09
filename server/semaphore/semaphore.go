package semaphore

import "sync"

type Semaphore struct {
	c  chan int
	wg sync.WaitGroup
}

func New(n int) *Semaphore {
	s := &Semaphore{
		c:  make(chan int, n),
		wg: sync.WaitGroup{},
	}
	return s
}

func (s *Semaphore) Acquire() {
	s.wg.Add(1)
	s.c <- 0
}

func (s *Semaphore) Release() {
	s.wg.Done()
	<-s.c
}

func (s *Semaphore) Wait() {
	s.wg.Wait()
}

func (s *Semaphore) GetLen() int {
	return len(s.c)
}
