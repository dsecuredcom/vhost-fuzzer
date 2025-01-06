package scanner

import (
	"bufio"
	"os"
	"strings"
)

const (
	batchSize            = 1000
	bufferSize           = 1024 * 1024 // 1MB buffer
	progressBatch        = 10000       // Update progress every 10K requests
	defaultIPChunkSize   = 10_000
	defaultHostChunkSize = 10_000
)

type BatchProcessor struct {
	ipFile     *os.File
	hostFile   *os.File
	paths      []string
	targetChan chan Target
	batchSize  int
}

func (bp *BatchProcessor) ProcessFilesChunked() error {
	defer close(bp.targetChan)

	// Wrap both files in bufio.Scanners
	ipScanner := bufio.NewScanner(bp.ipFile)
	ipScanner.Buffer(make([]byte, bufferSize), bufferSize)

	for {
		// 1) Read a chunk of IPs (up to defaultIPChunkSize)
		ipChunk, err := readChunk(ipScanner, defaultIPChunkSize)
		if err != nil {
			return err
		}
		if len(ipChunk) == 0 {
			// No more IPs left
			break
		}

		// 2) For each chunk of IPs, we need to re‐scan the hosts file from the beginning
		if _, err := bp.hostFile.Seek(0, 0); err != nil {
			return err
		}
		hostScanner := bufio.NewScanner(bp.hostFile)
		hostScanner.Buffer(make([]byte, bufferSize), bufferSize)

		for {
			// 3) Read a chunk of hosts (up to defaultHostChunkSize)
			hostChunk, err := readChunk(hostScanner, defaultHostChunkSize)
			if err != nil {
				return err
			}
			if len(hostChunk) == 0 {
				// No more hosts left
				break
			}

			// 4) Emit cross product of IP chunk and host chunk (plus all paths)
			for _, ip := range ipChunk {
				for _, host := range hostChunk {
					for _, path := range bp.paths {
						bp.targetChan <- Target{
							IP:       ip,
							Hostname: host,
							Path:     path,
						}
					}
				}
			}
		}
	}
	return nil
}

func readChunk(scanner *bufio.Scanner, chunkSize int) ([]string, error) {
	var lines []string
	for len(lines) < chunkSize && scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	return lines, scanner.Err()
}

func NewBatchProcessor(ipPath, hostPath string, paths []string, targetChan chan Target) (*BatchProcessor, error) {
	ipFile, err := os.Open(ipPath)
	if err != nil {
		return nil, err
	}

	hostFile, err := os.Open(hostPath)
	if err != nil {
		ipFile.Close()
		return nil, err
	}

	return &BatchProcessor{
		ipFile:     ipFile,
		hostFile:   hostFile,
		paths:      paths,
		targetChan: targetChan,
		batchSize:  batchSize,
	}, nil
}

func (bp *BatchProcessor) Close() {
	bp.ipFile.Close()
	bp.hostFile.Close()
}

func (bp *BatchProcessor) ProcessFiles() error {
	defer close(bp.targetChan)

	ipScanner := bufio.NewScanner(bp.ipFile)
	ipScanner.Buffer(make([]byte, bufferSize), bufferSize)

	for ipScanner.Scan() {
		ip := ipScanner.Text()
		if ip == "" {
			continue
		}

		if err := bp.processIPWithHosts(ip); err != nil {
			return err
		}
	}

	return ipScanner.Err()
}

func (bp *BatchProcessor) processIPWithHosts(ip string) error {
	_, err := bp.hostFile.Seek(0, 0)
	if err != nil {
		return err
	}

	hostScanner := bufio.NewScanner(bp.hostFile)
	hostScanner.Buffer(make([]byte, bufferSize), bufferSize)

	for hostScanner.Scan() {
		host := hostScanner.Text()
		if host == "" {
			continue
		}

		for _, path := range bp.paths {
			bp.targetChan <- Target{
				IP:       ip,
				Hostname: host,
				Path:     path,
			}
		}
	}

	return hostScanner.Err()
}
