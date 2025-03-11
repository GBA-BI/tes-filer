package s3

import (
	"context"
	"io"
	"net/http"
	"sync"

	"golang.org/x/time/rate"
)

type rateLimitingTransport struct {
	upLimiter   *rate.Limiter
	downLimiter *rate.Limiter
	transport   http.RoundTripper
}

func (t *rateLimitingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Body != nil {
		req.Body = &readLimiter{
			reader:  req.Body,
			limiter: t.upLimiter,
		}
	}

	resp, err := t.transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	resp.Body = &readLimiter{
		reader:  resp.Body,
		limiter: t.downLimiter,
	}

	return resp, nil
}

type readLimiter struct {
	reader  io.ReadCloser
	limiter *rate.Limiter
	mu      sync.RWMutex
}

func (r *readLimiter) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	if err != nil {
		return n, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	if err := r.limiter.WaitN(context.TODO(), n); err != nil {
		return 0, err
	}

	return n, nil
}

func (r *readLimiter) Close() error {
	return r.reader.Close()
}

func (r *readLimiter) UpdateRateLimit(newRate int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.limiter.SetLimit(rate.Limit(newRate))
}
