package scanner

import (
	"fmt"
	"github.com/schollz/progressbar/v3"
	"strings"
	"sync"

	"github.com/dsecuredcom/vhost-fuzzer/pkg/config"
	"github.com/fatih/color"
	"github.com/valyala/fasthttp"
)

type Scanner struct {
	config      *config.Config
	clients     *ClientPool
	targets     []Target
	results     chan Result
	rateLimiter chan struct{}
}

func (s *Scanner) Results() chan Result {
	return s.results
}

func (s *Scanner) CreateTargetChannel(ips, hosts []string, paths []string) (chan Target, int) {
	totalTargets := len(ips) * len(hosts) * len(paths)
	targetChan := make(chan Target, totalTargets)

	for _, ip := range ips {
		for _, host := range hosts {
			for _, path := range paths {
				if !strings.HasPrefix(path, "/") {
					path = "/" + path
				}
				targetChan <- Target{IP: ip, Hostname: host, Path: path}
			}
		}
	}
	close(targetChan)

	return targetChan, totalTargets
}

func NewScanner(cfg *config.Config) *Scanner {
	return &Scanner{
		config:      cfg,
		clients:     NewClientPool(),
		results:     make(chan Result, cfg.Concurrency*2),
		rateLimiter: make(chan struct{}, cfg.Concurrency),
	}
}

func (s *Scanner) Worker(targets <-chan Target, wg *sync.WaitGroup) {
	defer wg.Done()

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	for target := range targets {
		// Rate limiting
		s.rateLimiter <- struct{}{}

		req.Reset()
		resp.Reset()
		reqURI := fmt.Sprintf("%s://%s%s", s.config.Protocol, target.IP, target.Path)
		req.SetRequestURI(reqURI)
		req.SetHost(target.Hostname)
		req.Header.SetUserAgent("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36")

		for k, v := range s.config.Headers {
			req.Header.Set(k, v)
		}

		req.Header.SetBytesKV([]byte("Connection"), []byte("close"))

		hc := s.clients.getClient(target.IP, s.config)
		err := hc.DoTimeout(req, resp, s.config.RequestTimeout)

		if err != nil {
			s.results <- Result{Target: target, Error: err}
			<-s.rateLimiter
			continue
		}

		respCopy := fasthttp.AcquireResponse()
		resp.CopyTo(respCopy)
		result := Result{Target: target, Response: respCopy}
		s.results <- result
		<-s.rateLimiter
	}
}

func (s *Scanner) ProcessResults(bar *progressbar.ProgressBar, total int) {
	processed := 0
	for result := range s.results {
		processed++
		bar.Add(1)

		if result.Error != nil {
			if s.config.Verbose {
				fmt.Printf("Error scanning %s%s with host %s: %v\n",
					result.Target.IP, result.Target.Path, result.Target.Hostname, result.Error)
			}
			continue
		}

		statusMatch := s.config.StatusCode == 0 || result.Response.StatusCode() == s.config.StatusCode
		bodyMatch := true

		if len(s.config.BodyIncludes) > 0 {
			body := strings.ToLower(string(result.Response.Body()))
			bodyMatch = false
			var matchedTerm string

			for _, term := range s.config.BodyIncludes {
				if strings.Contains(body, strings.ToLower(term)) {
					bodyMatch = true
					matchedTerm = term
					if s.config.Verbose {
						fmt.Printf("Found matching term '%s' in response body\n", term)
					}
					break
				}
			}

			if bodyMatch {
				result.MatchedTerm = matchedTerm
			}
		}

		if !statusMatch || !bodyMatch {
			if s.config.Verbose {
				fmt.Printf("No match - Status match: %v, Body match: %v\n", statusMatch, bodyMatch)
			}
			continue
		}

		// Get the response body and truncate to 750 characters if needed
		body := strings.ReplaceAll(strings.ReplaceAll(string(result.Response.Body()), "\n", ""), "\r", "")
		if len(body) > 750 {
			body = body[:750] + "..."
		}

		green := color.New(color.FgGreen).SprintfFunc()
		matchInfo := ""
		if result.MatchedTerm != "" {
			matchInfo = fmt.Sprintf(" (matched term: '%s')", result.MatchedTerm)
		}

		fmt.Printf("\n%s\nBody preview: %s\n",
			green("Match found: %s%s with host %s (Status: %d)%s",
				result.Target.IP, result.Target.Path, result.Target.Hostname,
				result.Response.StatusCode(), matchInfo),
			body)

		if result.Response != nil {
			fasthttp.ReleaseResponse(result.Response)
		}

		if processed >= total {
			break
		}
	}
}
