package bot

import (
	"path"

	"github.com/go-faster/errors"
	"github.com/go-resty/resty/v2"
	"github.com/nktknshn/go-twitter-download-bot/cli/logging"
	"github.com/nktknshn/go-twitter-download-bot/twitter"
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

type Downloaded struct {
	Path   string
	Entity Downloadable
}

func (d Downloaded) IsPhoto() bool {
	_, ok := d.Entity.(twitter.Photo)
	return ok
}

func (d Downloaded) IsVideo() bool {
	_, ok := d.Entity.(twitter.VideoVariant)
	return ok
}

func (d *Downloader) Filename(td *twitter.TweetData, withfn interface{ Filename() string }) string {
	return td.Url.User + "_" + td.Url.ID + "_" + withfn.Filename()
}

type Downloadable interface {
	Filename() string
	URL() string
}

func (d *Downloader) DownloadTweetData(td *twitter.TweetData, destDir string) ([]Downloaded, error) {

	var toDownload = make([]Downloadable, 0, 4)
	var downloads []Downloaded

	for _, p := range td.Photos {
		toDownload = append(toDownload, p)
	}

	for _, v := range td.Videos {
		best, ok := v.Variants.VideoBestBitrate()
		if !ok {
			continue
		}
		toDownload = append(toDownload, best)
	}

	for _, p := range toDownload {
		path := path.Join(destDir, d.Filename(td, p))
		if err := d.Download(p.URL(), path); err != nil {
			return nil, errors.Wrap(err, "failed to download photo")
		}
		downloads = append(downloads, Downloaded{Path: path, Entity: p})
	}

	return downloads, nil
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
