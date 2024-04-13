package bot

import (
	"context"
	"path"

	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/telegram/message/unpack"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
)

type Filetype string

const (
	FiletypePhoto Filetype = "photo"
	FiletypeVideo Filetype = "video"
)

func (h *Handler) Upload(ctx context.Context, filepaths []string, filetype Filetype, title string, user tg.InputPeerClass) (*tg.Message, error) {
	if filetype == FiletypePhoto {
		return h.SendPhotos(ctx, filepaths, title, user)
	} else if filetype == FiletypeVideo {
		return h.SendVideos(ctx, filepaths, title, user)
	}

	return nil, errors.New("unknown filetype")

}

func (h *Handler) uploader() *uploader.Uploader {
	return uploader.NewUploader(h.Api)
}

func (h *Handler) uploaderWithSender() (*uploader.Uploader, *message.Sender) {
	uploader := h.uploader()
	sender := h.Sender.WithUploader(uploader)
	return uploader, sender
}

func (h *Handler) SendPhotos(ctx context.Context, filepaths []string, title string, user tg.InputPeerClass) (*tg.Message, error) {
	uploader, sender := h.uploaderWithSender()

	if len(filepaths) == 0 {
		return nil, errors.New("no files to upload")
	}

	uploads := make([]tg.InputFileClass, len(filepaths))

	for i, path := range filepaths {
		u, err := uploader.FromPath(ctx, path)

		if err != nil {
			return nil, errors.Wrap(err, "failed to upload")
		}

		uploads[i] = u
	}

	if len(uploads) == 1 {
		doc := message.UploadedPhoto(uploads[0], styling.Plain(title))
		msg, err := unpack.Message(sender.To(user).Media(ctx, doc))
		if err != nil {
			return nil, errors.Wrap(err, "failed to send")
		}
		return msg, nil
	}

	builders := make([]message.MultiMediaOption, len(uploads))

	for i, upload := range uploads {
		builders[i] = message.UploadedPhoto(upload, styling.Plain(title))
	}

	msg, err := unpack.Message(sender.To(user).Album(ctx, builders[0], builders[1:]...))

	if err != nil {
		return nil, errors.Wrap(err, "failed to send")
	}

	return msg, nil
}

func (h *Handler) SendVideos(ctx context.Context, filepaths []string, title string, user tg.InputPeerClass) (*tg.Message, error) {
	uploader, sender := h.uploaderWithSender()
	u, err := uploader.FromPath(ctx, filepaths[0])

	if err != nil {
		return nil, errors.Wrap(err, "failed to upload")
	}

	doc := message.UploadedDocument(u, styling.Plain(title))
	doc.Filename(path.Base(filepaths[0]))
	doc.MIME("video/mp4")

	msg, err := unpack.Message(sender.To(user).Media(ctx, doc))

	if err != nil {
		return nil, errors.Wrap(err, "failed to send")
	}
	return msg, nil
}
