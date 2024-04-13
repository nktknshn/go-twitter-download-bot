package bot

import (
	"context"
	"path"

	"github.com/go-faster/errors"
	"github.com/gotd/td/crypto"
	"github.com/gotd/td/telegram/message/unpack"
	"github.com/gotd/td/tg"
	"go.uber.org/zap"
)

// upload using raw telegram api. For science
func (h *Handler) uploadDownloadsRaw(ctx context.Context, peer tg.InputPeerClass, caption string, downloads []Downloaded) error {

	uploader, _ := h.uploaderWithSender()

	multiMedia := make([]tg.InputSingleMedia, len(downloads))

	for i, d := range downloads {

		// upload chunks
		u, err := uploader.FromPath(ctx, d.Path)
		if err != nil {
			return errors.Wrap(err, "failed to upload")
		}

		rnd, _ := crypto.RandInt64(crypto.DefaultRand())

		// doc := message.UploadedPhoto(u)

		if d.IsPhoto() {
			h.Logger.Info("Adding photo", zap.String("path", d.Path))

			// attach media to a peer creating ID, FileReference and AccessHash
			// https://core.telegram.org/method/messages.uploadMedia
			// Upload a file and associate it to a chat (without actually sending it to the chat)
			media, err := h.Api.MessagesUploadMedia(ctx, &tg.MessagesUploadMediaRequest{
				Peer: peer,
				// file uploaded in chunks
				Media: &tg.InputMediaUploadedPhoto{File: u},
			})

			if err != nil {
				h.Logger.Error("failed to upload media", zap.Error(err))
				return errors.Wrap(err, "failed to upload media")
			}

			mediaPhoto, ok := media.(*tg.MessageMediaPhoto)

			if !ok {
				h.Logger.Error("failed to MessageMediaPhoto", zap.Any("media", media))
				return errors.New("failed to MessageMediaPhoto")
			}

			photo, ok := mediaPhoto.Photo.AsNotEmpty()

			if !ok {
				h.Logger.Error("failed to AsNotEmpty", zap.Any("media", media))
				return errors.New("failed to AsNotEmpty")
			}

			multiMedia[i] = tg.InputSingleMedia{
				Media:    &tg.InputMediaPhoto{ID: photo.AsInput()},
				RandomID: rnd,
			}
		} else if d.IsVideo() {
			h.Logger.Info("Adding video", zap.String("path", d.Path))

			media, err := h.Api.MessagesUploadMedia(ctx, &tg.MessagesUploadMediaRequest{
				Peer: peer,
				Media: &tg.InputMediaUploadedDocument{
					File:     u,
					MimeType: "video/mp4",
					Attributes: []tg.DocumentAttributeClass{
						&tg.DocumentAttributeVideo{},
						&tg.DocumentAttributeFilename{FileName: path.Base(d.Path)},
					},
				},
			})

			if err != nil {
				h.Logger.Error("failed to upload media", zap.Error(err))
				return errors.Wrap(err, "failed to upload media")
			}

			mediaDocument, ok := media.(*tg.MessageMediaDocument)

			if !ok {
				h.Logger.Error("failed to get MessageMediaDocument", zap.Any("media", media))
				return errors.New("failed to get MessageMediaDocument")
			}

			doc, ok := mediaDocument.Document.AsNotEmpty()

			if !ok {
				h.Logger.Error("failed to AsNotEmpty", zap.Any("media", media))
				return errors.New("failed to AsNotEmpty")
			}

			multiMedia[i] = tg.InputSingleMedia{
				Media: &tg.InputMediaDocument{
					ID: doc.AsInput(),
				},
				RandomID: rnd,
			}
		} else {
			h.Logger.Error("unsupported media type", zap.Any("download", d))
			return errors.New("unsupported media type")
		}
	}

	req := &tg.MessagesSendMultiMediaRequest{
		MultiMedia: multiMedia,
		Peer:       peer,
	}

	h.Logger.Info("Sending media", zap.Int("count", len(multiMedia)), zap.Any("req", req))

	uploadedMessage, err := unpack.Message(h.Api.MessagesSendMultiMedia(ctx, req))

	if err != nil {
		return errors.Wrap(err, "failed to send message")
	}

	_ = uploadedMessage

	return nil
}
