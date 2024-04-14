package bot

import (
	"context"
	"time"

	"github.com/go-faster/errors"
	"github.com/gotd/contrib/middleware/floodwait"
	"github.com/gotd/contrib/middleware/ratelimit"
	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"github.com/nktknshn/go-twitter-download-bot/cli/logging"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

func Run(ctx context.Context,
	downloadFolder string,
	opts ...option,
) error {

	options := &options{
		logger:         logging.GetLogger().Named("bot"),
		forwardTo:      0,
		useRateLimiter: true,
		debugTelegram:  false,
		sessionFile:    "twitter-downloader-session.json",
	}

	for _, opt := range opts {
		opt(options)
	}

	handler := &Handler{
		Logger:            options.logger,
		Dispatcher:        tg.NewUpdateDispatcher(),
		DownloadFolder:    downloadFolder,
		AdminID:           options.adminID,
		RestrictToAdminID: options.restrictToAdmin,
		ForwardTo:         options.forwardTo,
		DebugTelegram:     options.debugTelegram,
		IncludeText:       options.includeText,
		IncludeURL:        options.includeURL,
		IncludeBotName:    options.includeBotName,
		LimitPerDay:       options.limitPerDay,
		LimitPending:      options.limitPending,
	}

	tgLogger := zap.NewNop()

	if handler.DebugTelegram {
		tgLogger = logging.GetLogger().Named("telegram")
	}

	waiter := floodwait.NewWaiter().WithCallback(func(ctx context.Context, wait floodwait.FloodWait) {
		// Notifying about flood wait.
		handler.Logger.Warn("Flood wait", zap.Duration("wait", wait.Duration))
	})

	middlewares := []telegram.Middleware{}

	if options.useRateLimiter {
		middlewares = append(middlewares, waiter)
		middlewares = append(middlewares, ratelimit.New(rate.Every(time.Millisecond*100), 5))
	}

	tgopts := telegram.Options{
		Logger:        tgLogger,
		UpdateHandler: handler,
		SessionStorage: &session.FileStorage{
			Path: options.sessionFile,
		},
		Device: telegram.DeviceConfig{
			DeviceModel:    "pc",
			SystemVersion:  "linux",
			AppVersion:     "0.0.1",
			SystemLangCode: "en",
		},
		Middlewares: middlewares,
	}

	runBot := func(ctx context.Context) error {
		return telegram.BotFromEnvironment(ctx, tgopts,
			func(ctx context.Context, client *telegram.Client) error {
				handler.Logger.Info("Setting up handler")
				return handler.Init(client)
			},
			func(ctx context.Context, client *telegram.Client) error {
				handler.Logger.Info("Connected")

				if err := handler.OnConnected(ctx); err != nil {
					return errors.Wrap(err, "failed to get self username")
				}

				return telegram.RunUntilCanceled(ctx, client)
			},
		)
	}

	if options.useRateLimiter {
		return waiter.Run(ctx, runBot)
	}

	return runBot(ctx)
}
