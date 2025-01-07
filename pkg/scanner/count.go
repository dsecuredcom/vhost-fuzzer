package scanner

import (
	"bufio"
	"fmt"
	"os"
)

func CountTotalTargets(ipsFile, hostsFile string, pathsCount int) (int64, error) {
	ipsCount, err := countLinesStreaming(ipsFile)
	if err != nil {
		return 0, fmt.Errorf("error counting IPs: %v", err)
	}
	hostsCount, err := countLinesStreaming(hostsFile)
	if err != nil {
		return 0, fmt.Errorf("error counting hosts: %v", err)
	}
	return int64(ipsCount * hostsCount * pathsCount), nil
}

func countLinesStreaming(filename string) (int, error) {
	file, err := os.Open(filename)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, bufferSize), bufferSize)
	count := 0
	for scanner.Scan() {
		if line := scanner.Text(); line != "" {
			count++
		}
	}
	if err := scanner.Err(); err != nil {
		return 0, err
	}
	return count, nil
}
