package scanner

import (
	"fmt"
	"strings"

	"github.com/valyala/fasthttp"
)

func (s *Scanner) checkTarget(target Target, req *fasthttp.Request, resp *fasthttp.Response) string {
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
		if s.config.Verbose {
			fmt.Printf("\n=== Error ===\n")
			fmt.Printf("Failed to execute request to %s: %v\n", reqURI, err)
			fmt.Printf("========================\n")
		}
		return ""
	}
	statusCode := resp.StatusCode()
	if s.config.Verbose {
		fmt.Printf("\n=== Request ===\n")
		fmt.Printf("URI: %s\n", reqURI)
		fmt.Printf("Host: %s\n", target.Hostname)
		fmt.Printf("Method: %s\n", string(req.Header.Method()))
		req.Header.VisitAll(func(k, v []byte) {
			fmt.Printf("%s: %s\n", string(k), string(v))
		})
		fmt.Printf("\n=== Response ===\n")
		fmt.Printf("Status: %d\n", statusCode)
		resp.Header.VisitAll(func(k, v []byte) {
			fmt.Printf("%s: %s\n", string(k), string(v))
		})
		if len(resp.Body()) > 0 {
			fmt.Printf("\nBody (truncated):\n%s\n", truncateString(string(resp.Body()), 1000))
		}
		fmt.Printf("========================\n")
	}
	if s.config.HTTPStatusIs != 0 && statusCode != s.config.HTTPStatusIs {
		return ""
	}
	if s.config.HTTPBodyIncludes != "" {
		if !strings.Contains(string(resp.Body()), s.config.HTTPBodyIncludes) {
			return ""
		}
	}
	return fmt.Sprintf("\n[+] Found match - IP: %s, Host: %s, Path: %s, Status: %d", target.IP, target.Hostname, target.Path, statusCode)
}

func truncateString(str string, maxLen int) string {
	if len(str) <= maxLen {
		return str
	}
	return str[:maxLen] + "..."
}
