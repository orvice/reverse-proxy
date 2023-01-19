package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"golang.org/x/exp/slog"
)

func main() {
	remote, err := url.Parse(os.Getenv("DST"))
	if err != nil {
		panic(err)
	}

	slog.Info("proxy info", "dst.host", remote.Host)

	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = remote.Scheme
		req.Host = remote.Host
		req.URL.Host = remote.Host
		req.Header.Set("Host", remote.Host)

		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
	}
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		slog.Error("proxy error",
			err,
			"url", r.URL.String())

	}

	http.HandleFunc("/", handler(proxy, remote))
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

func handler(p *httputil.ReverseProxy, u *url.URL) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("proxy url",
			"url", r.URL.String())
		w.Header().Set("Host", u.Host)
		p.ServeHTTP(w, r)
	}
}
