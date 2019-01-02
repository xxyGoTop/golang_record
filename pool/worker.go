package Pool

import (
	"time"
)

//执行任务
type Worker struct {
	pool *Pool

	//任务list
	task chan f
	//更新时间
	recycleTime time.Time
}

//启动一个协程来执行任务
func (w *Worker) run() {
	go func() {
		for f := range w.task {
			if f == nil {
				w.pool.decRunning()
				return
			}
			f()
			w.pool.putWorker(w)
		}
	}()
}
