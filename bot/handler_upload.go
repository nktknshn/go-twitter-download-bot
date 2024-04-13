package bot

import (
	"context"
	"path"

	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/telegram/uploader"
	"go.uber.org/zap"
)

func (h *Handler) uploader() *uploader.Uploader {
	return uploader.NewUploader(h.Api)
}

func (h *Handler) uploaderWithSender() (*uploader.Uploader, *message.Sender) {
	uploader := h.uploader()
	sender := h.Sender.WithUploader(uploader)
	return uploader, sender
}

func (h *Handler) uploadDownloads(ctx context.Context, downloads []Downloaded, caption string) ([]message.MultiMediaOption, error) {

	uploader, _ := h.uploaderWithSender()
	uploads := make([]message.MultiMediaOption, len(downloads))

	for i, download := range downloads {
		h.Logger.Info("Uploading media", zap.String("path", download.Path))

		u, err := uploader.FromPath(ctx, download.Path)
		if err != nil {
			return nil, errors.Wrap(err, "upload media")
		}

		st := []styling.StyledTextOption{}

		if i == 0 {
			// https://core.telegram.org/api/files#albums-grouped-media
			// For photo albums, clients should display an album caption only if exactly one photo in the group has a caption, otherwise no album caption should be displayed, and only when viewing in detail a specific photo of the group the caption should be shown.
			// so we add caption to the first message
			st = []styling.StyledTextOption{styling.Plain(caption)}
		}

		if download.IsPhoto() {
			uploads[i] = message.UploadedPhoto(u, st...)
		} else if download.IsVideo() {
			doc := message.UploadedDocument(u, st...)
			doc.Filename(path.Base(download.Path))
			doc.MIME("video/mp4")
			uploads[i] = doc
		} else {
			return nil, errors.New("unsupported media type")
		}
	}

	return uploads, nil
}
