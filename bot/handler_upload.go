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

func (h *Handler) Upload(ctx context.Context, filepath string, filetype Filetype, title string, user tg.InputPeerClass) (*tg.Message, error) {
	if filetype == FiletypePhoto {
		return h.SendPhoto(ctx, filepath, title, user)
	} else if filetype == FiletypeVideo {
		return h.SendVideo(ctx, filepath, title, user)
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

func (h *Handler) SendPhoto(ctx context.Context, filepath, title string, user tg.InputPeerClass) (*tg.Message, error) {
	uploader, sender := h.uploaderWithSender()
	u, err := uploader.FromPath(ctx, filepath)

	if err != nil {
		return nil, errors.Wrap(err, "failed to upload")
	}
	doc := message.UploadedPhoto(u, styling.Plain(title))
	msg, err := unpack.Message(sender.To(user).Media(ctx, doc))

	if err != nil {
		return nil, errors.Wrap(err, "failed to send")
	}
	return msg, nil
}

func (h *Handler) SendVideo(ctx context.Context, filepath, title string, user tg.InputPeerClass) (*tg.Message, error) {
	uploader, sender := h.uploaderWithSender()
	u, err := uploader.FromPath(ctx, filepath)

	if err != nil {
		return nil, errors.Wrap(err, "failed to upload")
	}

	doc := message.UploadedDocument(u, styling.Plain(title))
	doc.Filename(path.Base(filepath))
	doc.MIME("video/mp4")

	msg, err := unpack.Message(sender.To(user).Media(ctx, doc))

	if err != nil {
		return nil, errors.Wrap(err, "failed to send")
	}
	return msg, nil
}
