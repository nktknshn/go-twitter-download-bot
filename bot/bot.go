package bot

import (
	"context"

	"github.com/go-faster/errors"
	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"github.com/nktknshn/go-twitter-fun/cli/logging"
	"go.uber.org/zap"
)

func Run(ctx context.Context,
	adminID int64,
	downloadFolder string,
	opts ...option,
) error {

	options := &options{
		logger:         logging.GetLogger().Named("bot"),
		forwardTo:      0,
		useRateLimiter: true,
		debugTelegram:  false,
	}

	for _, opt := range opts {
		opt(options)
	}

	handler := &Handler{
		Logger:         options.logger,
		Dispatcher:     tg.NewUpdateDispatcher(),
		DownloadFolder: downloadFolder,
		AdminID:        adminID,
		UploadTo:       options.forwardTo,
		DebugTelegram:  options.debugTelegram,
	}

	tgLogger := zap.NewNop()
	if handler.DebugTelegram {
		tgLogger = logging.GetLogger().Named("telegram")
	}

	tgopts := telegram.Options{
		Logger:        tgLogger,
		UpdateHandler: handler,
		SessionStorage: &session.FileStorage{
			Path: "twitter-downloader-session.json",
		},
		Device: telegram.DeviceConfig{
			DeviceModel:    "pc",
			SystemVersion:  "linux",
			AppVersion:     "0.0.1",
			SystemLangCode: "en",
		},
	}

	err := telegram.BotFromEnvironment(context.Background(), tgopts,
		func(ctx context.Context, client *telegram.Client) error {
			handler.Logger.Info("Setting up handler")
			handler.Init(client)
			return nil
		},
		func(ctx context.Context, client *telegram.Client) error {
			handler.Logger.Info("Connected")

			if err := handler.InitSelfUsername(ctx); err != nil {
				return errors.Wrap(err, "failed to get self username")
			}

			if handler.UploadTo != 0 {
				if err := handler.InitChannelAccessHash(ctx); err != nil {
					return errors.Wrap(err, "failed to get channel access hash")
				}
			}

			return telegram.RunUntilCanceled(ctx, client)
		},
	)

	return err
}
