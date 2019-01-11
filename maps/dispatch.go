package maps

import (
	"sync"
	"encoding/json"
	"fmt"
)


const (
	moreData int = iota
	done
)

type Connection interface {
	Send([]byte) error
	Receive() ([]byte, error)
}

type Worker struct {
	connection Connection
	maps []string
	sync.Mutex
}

func (w *Worker) doCommand(m NetworkMapper, d Keyed, k chan<- Keyed) error{
	w.Lock()
	defer w.Unlock()
	a, err := json.Marshal(&command{
		ID: m.ID(),
		Data: d,
	})
	if err != nil {
		return err
	}
	err = w.connection.Send(a)
	if err != nil {
		return err
	}
	for {
		data, err := w.connection.Receive()
		if err != nil {
			return err
		}
		o := &response{
			Data: m.NewOut(),
		}
		err = json.Unmarshal(data, o)
		if err != nil {
			return err
		}
		if o.ResponseCode == done {
			return nil
		}
		k <- o.Data
	}
}

func (s *Source) MapDispatch(m NetworkMapper) *Source {
	newChan := &Source{
		channel: make(chan Keyed, 100),
		w:       s.w,
	}
	go func() {
		wg := sync.WaitGroup{}
		for x := range s.channel {
			wg.Add(1)
			go func(x Keyed) {
				var err error
				for {
					id := m.ID()
					worker := s.w.getAndRemoveWorker(id, x.Key())
					fmt.Println("Dispatching")
					err = worker.doCommand(m, x, newChan.channel)
					if err != nil {
						fmt.Println("didn't work, repeating")
						continue
					}
					fmt.Println("Finished Job")
					s.w.FinishWorking(worker)
					break
				}
				wg.Done()
			}(x)
		}
		wg.Wait()
		close(newChan.channel)
	}()
	return newChan
}

type WorkerPool struct {
	sync.RWMutex
	pool map [string] *actionPool
	getWorkerLock sync.Mutex
}

func NewWorkerPool() *WorkerPool{
	return &WorkerPool{
		pool: make(map[string] *actionPool),
	}
}

func (w *WorkerPool) Register(maps ...NetworkMapper) {
	w.Lock()
	for _, nm := range maps {
		w.pool[nm.ID()] = newActionPool()
	}
	w.Unlock()
}

func (w *WorkerPool) RemoveWorker(worker *Worker) {
	w.RLock()
	for _, mapper := range worker.maps {
		w.pool[mapper].removeWorker(worker)
	}
	w.RUnlock()
}

func (w *WorkerPool) getAndRemoveWorker(mapperID string, keyedID int) *Worker{
	w.RLock()
	w.getWorkerLock.Lock()
	pool := w.pool[mapperID]
	worker := pool.getWorker(keyedID)
	for _, mapper := range worker.maps {
		w.pool[mapper].removeWorker(worker)
	}
	w.RUnlock()
	w.getWorkerLock.Unlock()
	return worker
}

func (w *WorkerPool) AddWorker(c Connection) {
	go func() {
		dat, err := c.Receive()
		if err != nil {
			return
		}
		begin := &beginMessage{}
		err = json.Unmarshal(dat, begin)
		if err != nil {
			return
		}
		worker := &Worker{
			connection: c,
			maps:       begin.Maps,
		}
		w.RLock()
		for _, mapstr := range worker.maps {
			w.pool[mapstr].addWorker(worker)
		}
		w.RUnlock()
	}()

}

func (w *WorkerPool) FinishWorking(worker *Worker) {
	w.RLock()
	for _, mapstr := range worker.maps {
		w.pool[mapstr].addWorker(worker)
	}
	w.RUnlock()
}

type actionPool struct {
	workers []*Worker
	waiters [] chan *Worker
	indexes map[*Worker] int
	sync.Mutex
}

func newActionPool() *actionPool {
	return &actionPool{
		indexes: make(map[*Worker] int),
	}
}

func (a *actionPool) getWorker(id int) *Worker {
	a.Lock()
	if len(a.workers) > 0 {
		defer a.Unlock()
		return a.workers[PosMod(id, len(a.workers))]
	} else {
		retchan := make(chan *Worker)
		a.waiters = append(a.waiters, retchan)
		a.Unlock()
		return <-retchan
	}
}

func (a *actionPool) addWorker(worker *Worker) {
	a.Lock()
	a.workers = append(a.workers, worker)
	a.indexes[worker] = len(a.workers) - 1
	if len(a.waiters) > 0 {
		waiter := a.waiters[0]
		a.waiters = a.waiters[1:]
		waiter <- worker
	}
	a.Unlock()
}

func (a *actionPool) removeWorker(worker *Worker) {
	a.Lock()
	idx := a.indexes[worker]
	a.workers[idx] = a.workers[len(a.workers)-1]
	a.indexes[a.workers[idx]] = idx
	a.workers = a.workers[:len(a.workers)-1]
	a.Unlock()
}

type command struct {
	ID string
	Data Keyed

}

type response struct {
	Data Keyed
	ResponseCode int
}

type commandID struct {
	ID string
}

type commandData struct {
	Data Keyed
}

type Host struct {
	sync.Mutex
	maps map[string]NetworkMapper
	idlist []string
	c Connection
}

type beginMessage struct {
	Maps []string
}

func NewHost(c Connection) *Host {
	return &Host{
		maps: make(map[string]NetworkMapper),
		c: c,
	}
}

func (h *Host) Register(maps ...NetworkMapper) {
	for _, m := range maps {
		h.Lock()
		h.maps[m.ID()] = m
		h.idlist = append(h.idlist, m.ID())
		h.Unlock()
	}
}

func (h *Host) Start() error {
	dat, err := json.Marshal(&beginMessage{
		Maps: h.idlist,
	})
	if err != nil {
		return err
	}
	err = h.c.Send(dat)
	if err != nil {
		return err
	}
	go func() {
		for {
			dat, err := h.c.Receive()
			fmt.Println("Received Job: ", string(dat))
			if err != nil {
				fmt.Println("Error receiving job data", err)

			}
			id := new(commandID)
			json.Unmarshal(dat, id)
			mapper, ok := h.maps[id.ID]
			if !ok {
				return
			}
			indata := &commandData{Data:mapper.NewIn()}
			json.Unmarshal(dat, indata)
			outchan := make(chan Keyed, 100)
			go func() {
				for x:= range outchan {
					fmt.Println("Starting send")
					respData, err := json.Marshal(&response{
						ResponseCode:moreData,
						Data:x,
					})
					if err != nil {
						fmt.Println(err)
						return
					}
					fmt.Println("sending", string(respData))
					err = h.c.Send(respData)
					if err != nil {
						fmt.Println("boi", err)
						return
					}
					fmt.Println("Sended")
				}
				respData, _ := json.Marshal(&response{
					ResponseCode:done,
					Data: mapper.NewOut(),
				})
				fmt.Println("Sending Done With Job", string(respData))
				err := h.c.Send(respData)
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Println("Finished Sending Done")
			}()
			mapper.Do(indata.Data, outchan)
			close(outchan)
		}
	}()
	return nil
}