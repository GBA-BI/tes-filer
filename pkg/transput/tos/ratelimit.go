package tos

import (
	"sync"
	"time"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"

	"github.com/GBA-BI/tes-filer/pkg/consts"
	"github.com/GBA-BI/tes-filer/pkg/log"
)

type downloadEventListenerAndRateLimiter struct {
	eventListenerAndRateLimiter
}

func newDownloadEventListenerAndRateLimiter(maxBandwidth, bandwidth int64, logger log.Logger) *downloadEventListenerAndRateLimiter {
	return &downloadEventListenerAndRateLimiter{
		eventListenerAndRateLimiter: *newListenerAndLimiter(maxBandwidth, bandwidth, logger),
	}
}

func (c *downloadEventListenerAndRateLimiter) EventChange(event *tos.DownloadEvent) {
	c.eventChange(event.Err)
}

type uploadEventListenerAndRateLimiter struct {
	eventListenerAndRateLimiter
}

func newUploadEventListenerAndRateLimiter(maxBandwidth, bandwidth int64, logger log.Logger) *uploadEventListenerAndRateLimiter {
	return &uploadEventListenerAndRateLimiter{
		eventListenerAndRateLimiter: *newListenerAndLimiter(maxBandwidth, bandwidth, logger),
	}
}

func (c *uploadEventListenerAndRateLimiter) EventChange(event *tos.UploadEvent) {
	c.eventChange(event.Err)
}

type eventListenerAndRateLimiter struct {
	tosRateLimiter      tos.RateLimiter
	maxBandwidth        int64
	bandwidth           int64
	lastSpeedDownTime   *time.Time
	lastSpeedChangeTime *time.Time
	hasNonRateLimitErr  bool
	hasRateLimitErr     bool
	lock                sync.Mutex
	logger              log.Logger
}

func newListenerAndLimiter(maxBandwidth, bandwidth int64, logger log.Logger) *eventListenerAndRateLimiter {
	if bandwidth == 0 {
		bandwidth = maxBandwidth
	}
	if maxBandwidth > 0 && bandwidth > maxBandwidth {
		bandwidth = maxBandwidth
	}

	return &eventListenerAndRateLimiter{
		tosRateLimiter: tos.NewDefaultRateLimit(bandwidth, bandwidth),
		maxBandwidth:   maxBandwidth,
		bandwidth:      bandwidth,
		logger:         logger,
	}
}

func (c *eventListenerAndRateLimiter) eventChange(err error) {
	if err == nil {
		return
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	if !isRateLimitError(err) {
		c.logger.Warnf("occur error: %v", err)
		c.hasNonRateLimitErr = true
		return
	}

	// rate limit occurs
	c.hasRateLimitErr = true

	// last speed down less than 10s, consider as multi events at the same speed, so no speed down
	if c.lastSpeedDownTime != nil && time.Since(*c.lastSpeedDownTime) < 10*time.Second {
		return
	}

	now := time.Now()
	c.lastSpeedDownTime = &now
	c.lastSpeedChangeTime = &now
	c.downBandwidth()
	c.tosRateLimiter = tos.NewDefaultRateLimit(c.bandwidth, c.bandwidth)
}

func (c *eventListenerAndRateLimiter) onlyOccurRateLimitErr() bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.hasRateLimitErr && !c.hasNonRateLimitErr
}

func (c *eventListenerAndRateLimiter) Acquire(want int64) (ok bool, timeToWait time.Duration) {
	c.lock.Lock()
	defer c.lock.Unlock()

	// no limit
	if c.bandwidth == 0 {
		return true, 0
	}
	// bandwidth is maxBandwidth, no more speed up
	if c.maxBandwidth > 0 && c.bandwidth == c.maxBandwidth {
		return c.tosRateLimiter.Acquire(want)
	}
	// stable less than 5min, no speed up
	if c.lastSpeedChangeTime != nil && time.Since(*c.lastSpeedChangeTime) < 5*time.Minute {
		return c.tosRateLimiter.Acquire(want)
	}

	now := time.Now()
	c.lastSpeedChangeTime = &now
	c.upBandwidth()
	c.tosRateLimiter = tos.NewDefaultRateLimit(c.bandwidth, c.bandwidth)
	if c.bandwidth == 0 {
		return true, 0
	}
	return c.tosRateLimiter.Acquire(want)
}

func (c *eventListenerAndRateLimiter) downBandwidth() {
	bandwidth := c.bandwidth
	if bandwidth == 0 {
		bandwidth = consts.DefaultMaxBandwidth
	}
	if bandwidth <= consts.DefaultMinBandwidth {
		c.logger.Infof("bandwidth is minimum %d", bandwidth)
	} else {
		bandwidth /= 2
		c.logger.Infof("bandwidth down to %d", bandwidth)
	}
	c.bandwidth = bandwidth
}

func (c *eventListenerAndRateLimiter) upBandwidth() {
	bandwidth := 2 * c.bandwidth
	if c.maxBandwidth == 0 && bandwidth > consts.DefaultMaxBandwidth {
		c.logger.Infof("no rate limit now")
		c.bandwidth = 0
		return
	}
	if c.maxBandwidth > 0 && bandwidth > c.maxBandwidth {
		bandwidth = c.maxBandwidth
	}
	c.logger.Infof("bandwidth up to %d", bandwidth)
	c.bandwidth = bandwidth
}
