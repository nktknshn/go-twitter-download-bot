package bot

import "go.uber.org/zap"

type options struct {
	logger         *zap.Logger
	debugTelegram  bool
	forwardTo      int64
	useRateLimiter bool
}

type option func(*options)

func WithLogger(logger *zap.Logger) option {
	return func(opts *options) {
		opts.logger = logger
	}
}

func WithDebugTelegram(debug bool) option {
	return func(opts *options) {
		opts.debugTelegram = debug
	}
}

func WithForwardTo(forwardTo int64) option {
	return func(opts *options) {
		opts.forwardTo = forwardTo
	}
}

func WithRateLimiter(useRateLimiter bool) option {
	return func(opts *options) {
		opts.useRateLimiter = useRateLimiter
	}
}
