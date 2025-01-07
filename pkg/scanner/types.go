package scanner

import "github.com/valyala/fasthttp"

type Target struct {
	IP       string
	Hostname string
	Path     string
}

type Result struct {
	Target      Target
	Response    *fasthttp.Response
	Error       error
	MatchedTerm string
}
