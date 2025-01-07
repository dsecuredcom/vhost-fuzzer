package config

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	IPsFile             string
	HostsFile           string
	Concurrency         int
	Paths               []string
	HTTPBodyIncludes    string
	HTTPStatusIs        int
	Verbose             bool
	RequestTimeout      time.Duration
	MaxIdleConnDuration time.Duration
	MaxConnDuration     time.Duration
	ReadTimeout         time.Duration
	WriteTimeout        time.Duration
	Protocol            string
	Headers             map[string]string
}

func ParseFlags() Config {

	var (
		config              Config
		pathsStr            string
		protocol            string
		requestTimeout      int
		maxIdleConnDuration int
		maxConnDuration     int
		readTimeout         int
		writeTimeout        int
		headersStr          string
	)

	flag.StringVar(&config.IPsFile, "ips", "", "File containing IP addresses")
	flag.StringVar(&config.HostsFile, "hosts", "", "File containing hostnames")
	flag.IntVar(&config.Concurrency, "concurrency", 100, "Number of concurrent requests")
	flag.StringVar(&pathsStr, "paths", "/", "Comma-separated list of paths to check")
	flag.StringVar(&protocol, "protocol", "http", "http/https")
	flag.StringVar(&config.HTTPBodyIncludes, "http-body-includes", "", "String to search for in response body")
	flag.IntVar(&config.HTTPStatusIs, "http-status-is", 0, "Expected HTTP status code")
	flag.IntVar(&requestTimeout, "request-timeout", 4, "Timeout for individual requests in seconds")
	flag.IntVar(&maxIdleConnDuration, "max-idle-timeout", 6, "Maximum idle connection duration in seconds")
	flag.IntVar(&maxConnDuration, "max-conn-timeout", 6, "Maximum connection duration in seconds")
	flag.IntVar(&readTimeout, "read-timeout", 5, "Read timeout in seconds")
	flag.IntVar(&writeTimeout, "write-timeout", 5, "Write timeout in seconds")
	flag.BoolVar(&config.Verbose, "verbose", false, "Show all requests and responses")
	flag.StringVar(&headersStr, "headers", "", "")

	flag.Parse()
	if config.IPsFile == "" || config.HostsFile == "" {
		flag.Usage()
		os.Exit(1)
	}
	config.RequestTimeout = time.Duration(requestTimeout) * time.Second
	config.MaxIdleConnDuration = time.Duration(maxIdleConnDuration) * time.Second
	config.MaxConnDuration = time.Duration(maxConnDuration) * time.Second
	config.ReadTimeout = time.Duration(readTimeout) * time.Second
	config.WriteTimeout = time.Duration(writeTimeout) * time.Second
	config.Paths = strings.Split(pathsStr, ",")
	for i, path := range config.Paths {
		if !strings.HasPrefix(path, "/") {
			config.Paths[i] = "/" + path
		}
	}
	protocol = strings.ToLower(protocol)
	if protocol != "http" && protocol != "https" {
		fmt.Printf("Unknown protocol '%s'. Falling back to http.\n", protocol)
		protocol = "http"
	}
	config.Protocol = protocol

	config.Headers = make(map[string]string)
	if headersStr != "" {
		parts := strings.Split(headersStr, ";")
		for _, part := range parts {
			p := strings.SplitN(strings.TrimSpace(part), ":", 2)
			if len(p) == 2 {
				hKey := strings.TrimSpace(p[0])
				hVal := strings.TrimSpace(p[1])
				if hKey != "" && hVal != "" {
					config.Headers[hKey] = hVal
				}
			}
		}
	}

	return config
}
