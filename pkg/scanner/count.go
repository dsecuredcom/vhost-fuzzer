package scanner

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func CountTotalTargetsWithFilter(ipsFile, hostsFile string,
	ipFilters, hostFilters []string, pathsCount int) (total int64, filtered int64, err error) {

	// Zähle erst alle IP-Zeilen und Host-Zeilen (ohne Filter).
	ipsCount, err := countLinesStreaming(ipsFile)
	if err != nil {
		return 0, 0, fmt.Errorf("error counting IPs: %v", err)
	}
	hostsCount, err := countLinesStreaming(hostsFile)
	if err != nil {
		return 0, 0, fmt.Errorf("error counting hosts: %v", err)
	}
	// Gesamtanzahl an Targets = (alle IPs) x (alle Hosts) x (Anzahl Paths)
	total = int64(ipsCount * hostsCount * pathsCount)

	// Nun zähle gefilterte IPs/Hosts:
	ipsFilteredCount, err := countLinesWithFilter(ipsFile, ipFilters)
	if err != nil {
		return 0, 0, fmt.Errorf("error counting filtered IPs: %v", err)
	}
	hostsFilteredCount, err := countLinesWithFilter(hostsFile, hostFilters)
	if err != nil {
		return 0, 0, fmt.Errorf("error counting filtered hosts: %v", err)
	}
	// Gefilterte Targets = (gefilterte IPs) x (gefilterte Hosts) x (Anzahl Paths)
	filtered = int64(ipsFilteredCount * hostsFilteredCount * pathsCount)

	return total, filtered, nil
}

// countLinesWithFilter ist ähnlich wie countLinesStreaming,
// wendet aber zusätzlich einen String-Filter an.
func countLinesWithFilter(filename string, filters []string) (int, error) {
	file, err := os.Open(filename)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, bufferSize), bufferSize)

	var count int
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if stringContainsAny(line, filters) {
			count++
		}
	}
	if err := scanner.Err(); err != nil {
		return 0, err
	}
	return count, nil
}

func stringContainsAny(s string, filters []string) bool {
	if len(filters) == 0 {
		// Keine Filter -> immer true
		return true
	}
	for _, f := range filters {
		if strings.Contains(s, f) {
			return true
		}
	}
	return false
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
