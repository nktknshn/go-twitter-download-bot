package bot

import (
	"github.com/go-faster/errors"
	"github.com/go-resty/resty/v2"
	"github.com/nktknshn/go-twitter-fun/cli/logging"
	"go.uber.org/zap"
)

type Downloader struct {
	httpClient *resty.Client
	Retries    int
	logger     *zap.Logger
}

func NewDownloader() *Downloader {
	return &Downloader{
		logger:     logging.GetLogger().Named("downloader"),
		httpClient: resty.New(),
		// could have used resty.New().SetRetryCount(3),
		Retries: 3,
	}
}

// path must include filename
func (d *Downloader) Download(url, path string) error {
	retries := d.Retries

	for {
		resp, err := d.httpClient.R().SetOutput(path).Get(url)

		if resp.IsSuccess() && err == nil {
			break
		}

		d.logger.Error("failed to download", zap.String("url", url), zap.String("path", path), zap.Error(err), zap.Int("retriesLeft", retries))

		if retries == 0 && err != nil {
			return errors.Wrap(err, "failed to download")
		}

		if retries == 0 && resp.IsError() {
			return errors.New("failed to download")
		}

		retries--
	}

	return nil
}
