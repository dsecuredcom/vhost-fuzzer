package scanner

import (
	"fmt"
	"sync"
	"time"

	"github.com/dsecuredcom/vhost-fuzzer/pkg/config"
	"github.com/schollz/progressbar/v3"
)

type Scanner struct {
	config         config.Config
	bar            *progressbar.ProgressBar
	targetChan     chan Target
	resultChan     chan string
	clients        *clientCache
	progressCount  int64
	progressMutex  sync.Mutex
	lastUpdateTime time.Time
}

func NewScanner(cfg config.Config, bar *progressbar.ProgressBar) *Scanner {
	return &Scanner{
		config:         cfg,
		bar:            bar,
		targetChan:     make(chan Target, cfg.Concurrency*2),
		resultChan:     make(chan string, cfg.Concurrency*2),
		clients:        newClientCache(),
		lastUpdateTime: time.Now(),
	}
}

func (s *Scanner) Run() {
	processor, err := NewBatchProcessor(
		s.config.IPsFile,
		s.config.HostsFile,
		s.config.Paths,
		s.targetChan,
		s.config.IPFilter,
		s.config.HostFilter,
	)

	if err != nil {
		fmt.Printf("Error initializing batch processor: %v\n", err)
		return
	}
	defer processor.Close()
	go func() {
		if err := processor.ProcessFilesChunked(); err != nil {
			fmt.Printf("Error processing files: %v\n", err)
		}
	}()
	pool := NewWorkerPool(s.config.Concurrency, s)
	pool.Start()
	done := make(chan struct{})
	go s.processResults(done)
	pool.Wait()
	close(s.resultChan)
	<-done
}

func (s *Scanner) updateProgress() {
	s.progressMutex.Lock()
	s.progressCount++
	if s.progressCount%progressBatch == 0 || time.Since(s.lastUpdateTime) > time.Second {
		s.bar.Set(int(s.progressCount))
		s.lastUpdateTime = time.Now()
	}
	s.progressMutex.Unlock()
}

func (s *Scanner) processResults(done chan struct{}) {
	for result := range s.resultChan {
		fmt.Println(result)
	}
	close(done)
}
