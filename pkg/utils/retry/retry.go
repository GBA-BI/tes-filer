package retry

import (
	"errors"
	"syscall"
	"time"

	"github.com/avast/retry-go/v4"

	"github.com/GBA-BI/tes-filer/pkg/log"
)

// MountTOSRetry retry when mounting tos and EIO occurs due to tos ratelimit
func MountTOSRetry(logger log.Logger, isMountTOS bool, fn func() error) error {
	if !isMountTOS {
		return fn()
	}
	return retry.Do(fn,
		retry.Attempts(5),
		retry.RetryIf(isIOError),
		retry.OnRetry(func(n uint, err error) {
			logger.Warnf("retry %d times with err %v, maybe due to tos ratelimit", n, err)
		}),
		retry.DelayType(retry.CombineDelay(retry.BackOffDelay, retry.RandomDelay)),
		retry.Delay(time.Second),
		retry.MaxDelay(time.Minute),
		retry.MaxJitter(time.Second*5),
		retry.LastErrorOnly(true),
	)
}

// when mounting tos, tos 429 will cause EIO
func isIOError(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, syscall.EIO)
}
