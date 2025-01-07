package scanner

import (
	"sync"

	"github.com/valyala/fasthttp"
)

type WorkerPool struct {
	workers     int
	scanner     *Scanner
	workerGroup sync.WaitGroup
}

func NewWorkerPool(workers int, scanner *Scanner) *WorkerPool {
	return &WorkerPool{
		workers: workers,
		scanner: scanner,
	}
}

func (wp *WorkerPool) Start() {
	reqPool := sync.Pool{
		New: func() interface{} {
			return &fasthttp.Request{}
		},
	}
	respPool := sync.Pool{
		New: func() interface{} {
			return &fasthttp.Response{}
		},
	}
	for i := 0; i < wp.workers; i++ {
		wp.workerGroup.Add(1)
		go wp.worker(&reqPool, &respPool)
	}
}

func (wp *WorkerPool) worker(reqPool, respPool *sync.Pool) {
	defer wp.workerGroup.Done()
	req := reqPool.Get().(*fasthttp.Request)
	resp := respPool.Get().(*fasthttp.Response)
	defer func() {
		reqPool.Put(req)
		respPool.Put(resp)
	}()
	for target := range wp.scanner.targetChan {
		if result := wp.scanner.checkTarget(target, req, resp); result != "" {
			wp.scanner.resultChan <- result
		}
		wp.scanner.updateProgress()
	}
}

func (wp *WorkerPool) Wait() {
	wp.workerGroup.Wait()
}
