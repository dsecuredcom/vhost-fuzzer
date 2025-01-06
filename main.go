package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/dsecuredcom/vhost-fuzzer/pkg/config"
	"github.com/dsecuredcom/vhost-fuzzer/pkg/scanner"
	"github.com/schollz/progressbar/v3"
)

func main() {
	// Parse configuration
	cfg := config.ParseFlags()

	// Set GOMAXPROCS to match concurrency
	if cfg.Concurrency > runtime.GOMAXPROCS(0) {
		runtime.GOMAXPROCS(cfg.Concurrency)
	}

	// Count total targets
	fmt.Println("[*] Counting targets...")
	startTime := time.Now()

	totalTargets, err := scanner.CountTotalTargets(cfg.IPsFile, cfg.HostsFile, len(cfg.Paths))
	if err != nil {
		fmt.Printf("[-] Error counting targets: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("[+] Found %d total targets (took %s)\n", totalTargets, time.Since(startTime))

	// Create progress bar with improved options
	bar := progressbar.NewOptions(
		int(totalTargets),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionShowBytes(false),
		progressbar.OptionSetWidth(30),
		progressbar.OptionShowCount(),
		progressbar.OptionSetDescription("[cyan][1/1][reset] Scanning targets..."),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "[green]=[reset]",
			SaucerHead:    "[green]>[reset]",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionOnCompletion(func() {
			fmt.Println("\n[+] Scan completed!")
		}),
	)

	// Start scanning
	scanner := scanner.NewScanner(cfg, bar)

	fmt.Printf("[*] Starting scan with %d workers...\n", cfg.Concurrency)
	scanStartTime := time.Now()

	scanner.Run()

	fmt.Printf("[+] Completed in %s\n", time.Since(scanStartTime))
}
