package bot

import (
	"context"
	"path"

	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/telegram/message/unpack"
	"github.com/gotd/td/tg"
	"github.com/nktknshn/go-twitter-download-bot/twitter"
	"go.uber.org/zap"
)

func (h *Handler) makeMessageText(td *twitter.TweetData) string {
	messageText := ""
	if h.includeText && td.CleanText() != "" {
		messageText += td.CleanText() + "\n"
	}
	if h.includeURL {
		messageText += td.Url.String() + "\n"
	}
	messageText += "@" + h.botName()
	return messageText
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

func (h *Handler) OnTwitterURLFromUser(ctx context.Context, entities tg.Entities, user *tg.PeerUser, m *tg.Message) error {

	h.Logger.Info("Received url", zap.String("url", m.Message))

	var workingMessage *tg.Message
	var err error

	if workingMessage, err = h.SendText(ctx, user, "Работаю... Working..."); err != nil {
		h.Logger.Error("failed to send message", zap.Error(err))
		return nil
	}

	defer h.removeMessage(ctx, workingMessage)

	td, err := h.twitter.GetTwitterData(ctx, m.Message)

	if err != nil {
		h.Logger.Error("failed to get twitter data", zap.Error(err))
		h.ReplyError(ctx, user, err, "Ошибка получения данных из твиттера. Error getting data from twitter.")
		return errors.Wrap(err, "get twitter data")
	}

	h.Logger.Debug("twitter data", zap.Any("data", td))

	downloads, err := h.downloader.DownloadTweetData(td, h.DownloadFolder)

	if err != nil {
		h.Logger.Error("failed to download tweet data", zap.Error(err))
		h.ReplyError(ctx, user, err, "Ошибка cкачки. Error downloading.")
		return errors.Wrap(err, "download tweet data")
	}

	messageText := h.makeMessageText(td)
	peer := h.inputUser(user)

	h.Logger.Info("Sending album", zap.Int("count", len(downloads)))

	uploads, err := h.uploadDownloads(ctx, downloads, messageText)

	if err != nil {
		h.Logger.Error("upload files", zap.Error(err))
		h.ReplyError(ctx, user, err, "Ошибка закачки в телеграм. Error uploading to telegram.")
		return errors.Wrap(err, "upload files")
	}

	sentMsg, err := unpack.Message(
		h.Sender.To(peer).
			Album(ctx, uploads[0], uploads[1:]...),
	)

	if err != nil {
		h.Logger.Error("send media group", zap.Error(err))
		h.ReplyError(ctx, user, err, "Ошибка отправки медигруппа в телеграм. Error sending media group.")
		return errors.Wrap(err, "send media group")
	}

	if h.ForwardTo == 0 {
		return nil
	}

	h.Logger.Info("Forwarding to channel", zap.Int64("channel", h.ForwardTo))

	_, err = h.Sender.To(h.inputChannelPeer()).
		ForwardMessages(h.inputUser(user), sentMsg).
		Send(ctx)

	if err != nil {
		h.Logger.Error("forward", zap.Error(err))
		return errors.Wrap(err, "forward")
	}

	return nil
}
