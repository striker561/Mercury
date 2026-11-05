package app

import (
	"io"
	"log"
	"net/http"
	"time"
)

const updaterRetryDelay = time.Second

// retryTransport retries idempotent updater requests once on transport errors,
// 429, or 5xx. The upstream provider reports the second response normally.
type retryTransport struct {
	base  http.RoundTripper
	delay time.Duration
}

func (t retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Method != http.MethodGet && req.Method != http.MethodHead {
		return t.roundTrip(req)
	}

	resp, err := t.roundTrip(req)
	if !updaterRetryable(resp, err) {
		return resp, err
	}

	if err != nil {
		if req.Context().Err() != nil {
			return nil, err
		}
		log.Printf("[mercury] updater request failed for %s: %v (retrying once)", updaterRequestURL(req), err)
	} else {
		log.Printf("[mercury] updater HTTP %d for %s (retrying once)", resp.StatusCode, updaterRequestURL(req))
		drainAndClose(resp)
	}

	delay := t.delay
	if delay == 0 {
		delay = updaterRetryDelay
	}
	select {
	case <-time.After(delay):
	case <-req.Context().Done():
		return nil, req.Context().Err()
	}

	retryResp, retryErr := t.roundTrip(req.Clone(req.Context()))
	if retryErr != nil {
		log.Printf("[mercury] updater request failed for %s after retry: %v", updaterRequestURL(req), retryErr)
	} else if retryResp.StatusCode >= 500 {
		log.Printf("[mercury] updater HTTP %d for %s after retry", retryResp.StatusCode, updaterRequestURL(req))
	}
	return retryResp, retryErr
}

func updaterRequestURL(req *http.Request) string {
	u := *req.URL
	u.RawQuery = ""
	u.Fragment = ""
	u.User = nil
	return u.String()
}

func (t retryTransport) roundTrip(req *http.Request) (*http.Response, error) {
	base := t.base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(req)
}

func updaterRetryable(resp *http.Response, err error) bool {
	if err != nil {
		return true
	}
	return resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500
}

func drainAndClose(resp *http.Response) {
	if resp.Body == nil {
		return
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
	_ = resp.Body.Close()
}
