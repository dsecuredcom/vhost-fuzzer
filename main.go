package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/schollz/progressbar/v3"

	"github.com/dsecuredcom/vhost-fuzzer/pkg/config"
	"github.com/dsecuredcom/vhost-fuzzer/pkg/scanner"
	"github.com/dsecuredcom/vhost-fuzzer/pkg/utils"
)

func main() {
	// Parse command line flags
	ipsFile := flag.String("ips", "", "File containing IP addresses")
	hostsFile := flag.String("hosts", "", "File containing hostnames")
	concurrency := flag.Int("concurrency", 100, "Number of concurrent workers")
	paths := flag.String("paths", "/", "Comma-separated list of paths to check")
	protocol := flag.String("protocol", "http", "Protocol to use (http/https)")
	bodyIncludes := flag.String("http-body-includes", "", "String to search for in response body")
	statusCode := flag.Int("http-status-is", 0, "Expected HTTP status code")
	requestTimeout := flag.Int("request-timeout", 10, "Timeout for individual requests in seconds")
	maxConnTimeout := flag.Int("max-conn-timeout", 5, "Maximum connection duration in seconds")
	readTimeout := flag.Int("read-timeout", 10, "Read timeout in seconds")
	headers := flag.String("headers", "", "Additional HTTP headers (Format: \"X1:v1;X2:v2\")")
	verbose := flag.Bool("verbose", false, "Show all requests and responses")

	flag.Parse()

	// Validate required flags
	if *ipsFile == "" || *hostsFile == "" {
		fmt.Println("Error: both -ips and -hosts flags are required")
		flag.Usage()
		os.Exit(1)
	}

	// Parse headers
	headerMap := make(map[string]string)
	if *headers != "" {
		headerPairs := strings.Split(*headers, ";")
		for _, pair := range headerPairs {
			parts := strings.SplitN(strings.TrimSpace(pair), ":", 2)
			if len(parts) == 2 {
				headerMap[parts[0]] = parts[1]
			}
		}
	}

	// Parse body includes terms
	bodyIncludeTerms := []string{}
	if *bodyIncludes != "" {
		for _, term := range strings.Split(*bodyIncludes, ",") {
			term = strings.TrimSpace(term)
			if term != "" {
				bodyIncludeTerms = append(bodyIncludeTerms, term)
			}
		}
	}

	// Create configuration
	cfg := &config.Config{
		Protocol:        *protocol,
		Concurrency:     *concurrency,
		RequestTimeout:  time.Duration(*requestTimeout) * time.Second,
		MaxConnDuration: time.Duration(*maxConnTimeout) * time.Second,
		ReadTimeout:     time.Duration(*readTimeout) * time.Second,
		WriteTimeout:    time.Duration(*readTimeout) * time.Second,
		Headers:         headerMap,
		BodyIncludes:    bodyIncludeTerms,
		StatusCode:      *statusCode,
		Verbose:         *verbose,
	}

	// Set GOMAXPROCS to match concurrency
	runtime.GOMAXPROCS(*concurrency)

	// Create scanner
	s := scanner.NewScanner(cfg)

	// Load targets
	ips := utils.LoadIpLines(*ipsFile)
	for _, ipRange := range ips {
		fmt.Printf("IP range: %s\n", ipRange)
	}
	syscall.Exit(1)
	hosts := utils.LoadHostLines(*hostsFile)
	pathList := strings.Split(*paths, ",")

	// Generate targets
	targetChan, totalTargets := s.CreateTargetChannel(ips, hosts, pathList)

	// Create progress bar
	bar := progressbar.Default(int64(totalTargets))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go s.Worker(targetChan, &wg)
	}

	// Process results
	go s.ProcessResults(bar, totalTargets)

	// Wait for completion
	wg.Wait()
	close(s.Results())

	fmt.Println("\nScan completed!")
}
