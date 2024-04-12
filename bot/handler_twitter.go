package bot

import (
	"context"
	"fmt"
	"path"

	"github.com/go-faster/errors"
	"github.com/gotd/td/tg"
	"go.uber.org/zap"
)

func (h *Handler) OnTwitterURLFromUser(ctx context.Context, entities tg.Entities, user *tg.PeerUser, m *tg.Message) error {

	h.Logger.Info("Received url", zap.String("url", m.Message))

	if _, err := h.ReplyText(ctx, user, "Работаю... Working..."); err != nil {
		h.Logger.Error("failed to send message", zap.Error(err))
		return nil
	}

	td, err := h.twitter.GetTwitterData(ctx, m.Message)

	if err != nil {
		h.Logger.Error("failed to get twitter data", zap.Error(err))
		h.ReplyError(ctx, user, err, "Ошибка получения данных из твиттера. Error getting data from twitter.")
		return errors.Wrap(err, "failed to get twitter data")
	}

	h.Logger.Debug("twitter data", zap.Any("data", td))

	// messageText := fmt.Sprintf("%s\n@%s", m.Message, h.botName())
	messageText := fmt.Sprintf("@%s", h.botName())

	var (
		filetype Filetype
		filepath string
		mediaURL string
	)

	if p, ok := td.Photo(); ok {
		filepath = path.Join(h.DownloadFolder, td.Url.User+"_"+td.Url.ID+".jpg")
		filetype = FiletypePhoto
		mediaURL = p.MediaURLHttps
	} else if v, ok := td.VideoBestBitrate(); ok {
		filepath = path.Join(h.DownloadFolder, td.Url.User+"_"+td.Url.ID+".mp4")
		filetype = FiletypeVideo
		mediaURL = v.URL
	} else {
		h.ReplyError(ctx, user, err, "В посте нет ни фото, ни видео. No photo or video found.")
		h.Logger.Error("no media found")
		return nil
	}

	h.Logger.Info("Downloading media", zap.String("url", mediaURL), zap.String("path", filepath))

	if err := h.downloader.Download(mediaURL, filepath); err != nil {
		h.ReplyError(ctx, user, err, "Ошибка закачки. Error downloading.")
		h.Logger.Error("failed to download", zap.Error(err))
		return errors.Wrap(err, "failed to download")
	}

	h.Logger.Info("Downloaded", zap.String("path", filepath))
	h.Logger.Info("Uploading media", zap.String("path", filepath))

	uploadedMessage, err := h.Upload(ctx, filepath, filetype, messageText, &tg.InputPeerUser{UserID: user.UserID})

	if err != nil {
		h.Logger.Error("failed to upload", zap.Error(err))
		h.ReplyError(ctx, user, err, "Ошибка закачки в телеграм. Error uploading to telegram.")
		return errors.Wrap(err, "failed to upload")
	}

	if h.UploadTo != 0 {
		h.Logger.Info("Forwarding to channel", zap.Int64("channel", h.UploadTo))

		_, err := h.Sender.To(h.inputChannelPeer()).
			ForwardMessages(h.inputUser(user), uploadedMessage).
			Send(ctx)

		if err != nil {
			h.Logger.Error("failed to forward", zap.Error(err))
			return errors.Wrap(err, "failed to forward")
		}
	}

	return nil
}
