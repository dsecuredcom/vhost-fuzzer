package config

import "time"

type Config struct {
	Protocol            string
	Concurrency         int
	RequestTimeout      time.Duration
	MaxConnDuration     time.Duration
	MaxIdleConnDuration time.Duration
	ReadTimeout         time.Duration
	WriteTimeout        time.Duration
	Headers             map[string]string
	BodyIncludes        []string
	StatusCode          int
	Verbose             bool
}
