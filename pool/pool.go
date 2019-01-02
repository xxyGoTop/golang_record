package Pool

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrPoolClosed = errors.New("this pool has been closed")
)

type sig struct{}

type f func() error

//协程池
type Pool struct {
	// pool容量
	capacity int32

	// goroutines/worker 运行数量
	running int32

	// 通知有可用的worker
	freeSignal chan sig

	// worker 协程
	workers []*Worker

	// release 释放标志
	release chan sig

	// lock mutex
	lock sync.Mutex

	// lock condtion
	cond *sync.Cond

	// lock mutex once
	once sync.Once
}

// 像这个pool提交个任务
func (p *Pool) Submit(task f) error {
	if len(p.release) > 0 {
		return ErrPoolClosed
	}
	p.getWorker().task <- task
	return nil
}

// 返回正在使用的worker个数
func (p *Pool) Running() int {
	return int(atomic.LoadInt32(&p.running))
}

// 返回pool的容量
func (p *Pool) Cap() int {
	return int(atomic.LoadInt32(&p.capacity))
}

//增加woker running 数量
func (p *Pool) incRunning() {
	atomic.AddInt32(&p.running, 1)
}

//减少worker running数量
func (p *Pool) decRunning() {
	atomic.AddInt32(&p.running, -1)
}

// 获取worker
func (p *Pool) getWorker() *Worker {
	var w *Worker
	waiting := false

	p.lock.Lock()
	defer p.lock.Unlock()

	freeWorkers := p.workers
	n := len(freeWorkers) - 1

	if n < 0 {
		waiting = p.Running() >= p.Cap()
	} else {
		w = freeWorkers[n]
		freeWorkers[n] = nil
		p.workers = freeWorkers[:n]
	}

	if waiting {
		for {
			p.cond.Wait()
			l := len(p.workers) - 1
			if l < 0 {
				continue
			}

			w = p.workers[l]
			p.workers[l] = nil
			p.workers = p.workers[:l]
			break
		}
	} else if w == nil {
		w = &Worker{
			pool: p,
			task: make(chan f, 1),
		}
		w.run()
		p.incRunning()
	}
	return w
}

// 回收worker
func (p *Pool) putWorker(worker *Worker) {
	worker.recycleTime = time.Now()

	p.lock.Lock()
	defer p.lock.Unlock()

	p.workers = append(p.workers, worker)
	p.cond.Signal()
}
