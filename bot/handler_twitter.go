package bot

import (
	"context"

	"github.com/go-faster/errors"
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

	sentMsgs, err := UnpackMultipleMessages(h.Sender.To(peer).
		Album(ctx, uploads[0], uploads[1:]...))

	if err != nil {
		h.Logger.Error("send media group", zap.Error(err))
		h.ReplyError(ctx, user, err, "Ошибка отправки медигруппы в телеграм. Error sending media group.")
		return errors.Wrap(err, "send media group")
	}

	if h.ForwardTo == 0 {
		return nil
	}

	sentMsgsIDs := make([]int, len(sentMsgs))

	for i, sentMsg := range sentMsgs {
		sentMsgsIDs[i] = sentMsg.ID
	}

	h.Logger.Info("Forwarding to channel", zap.Int64("channel", h.ForwardTo))

	_, err = h.Sender.To(h.inputChannelPeer()).
		ForwardIDs(h.inputUser(user), sentMsgsIDs[0], sentMsgsIDs[1:]...).
		Send(ctx)

	if err != nil {
		h.Logger.Error("forward", zap.Error(err))
		return errors.Wrap(err, "forward")
	}

	return nil
}
