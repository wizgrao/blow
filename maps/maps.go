package maps

import (
	"fmt"
	"sync"
)

type Generator interface {
	Do(chan<- Keyed)
}

type Mapper interface {
	Do(Keyed, chan<- Keyed)
}

type NetworkMapper interface {
	Mapper
	ID() string
	NewIn() Keyed // Empty struct for demarshalling
	NewOut() Keyed
}

type PrintMapper struct {
}




func (*PrintMapper) Do(k Keyed, c chan<- Keyed) {
	fmt.Printf("%+v\n", k)
	c <- k
}

type Source struct {
	channel chan Keyed
	w       *WorkerPool
}

func GeneratorSource(g Generator, pool *WorkerPool) *Source {
	newChan := &Source{
		channel: make(chan Keyed, 100),
		w:       pool,
	}
	go func() {
		g.Do(newChan.channel)
		close(newChan.channel)
	}()
	return newChan
}

func (s *Source) Sink() {
	for range s.channel {
	}
}

func (s *Source) MapLocal(m Mapper) *Source {
	newChan := &Source{
		channel: make(chan Keyed, 100),
		w:       s.w,
	}
	go func() {
		for x := range s.channel {
			m.Do(x, newChan.channel)
		}
		close(newChan.channel)
	}()
	return newChan
}

func (s *Source) MapLocalParallel(m Mapper, channels int) *Source {
	wg := sync.WaitGroup{}
	newChan := &Source{
		channel: make(chan Keyed, 100),
		w:       s.w,
	}
	inchans := make([]chan Keyed, channels)
	for i := range inchans {
		wg.Add(1)
		inchans[i] = make(chan Keyed, 100)
		go func(c chan Keyed) {
			for x := range c {
				m.Do(x, newChan.channel)
			}
			wg.Done()
		}(inchans[i])
	}

	go func() {
		for x := range s.channel {
			inchans[PosMod(x.Key(), channels)] <- x
		}
		for _, c := range inchans {
			close(c)
		}
		wg.Wait()
		close(newChan.channel)
	}()

	return newChan
}

type Keyed interface {
	Key() int //should be hash
}



func PosMod(x, y int) int {
	if x < 0 {
		x *= -1
	}
	return x % y
}

