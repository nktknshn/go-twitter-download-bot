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
	if h.IncludeText && td.CleanText() != "" {
		messageText += td.TweetText() + "\n"
	}
	if h.IncludeURL {
		messageText += td.Url.String() + "\n"
	}
	if h.IncludeBotName {
		messageText += "@" + h.botName()
	}
	return messageText
}

func (h *Handler) onTwitterURLFromUser(ctx context.Context, entities tg.Entities, user *tg.PeerUser, m *tg.Message) error {

	h.Logger.Info("Received url", zap.String("url", m.Message))

	h.updateQueryCountLimit(user.UserID)

	cq, cqr := h.canQuery(user.UserID)

	if !cq && cqr == reasonLimit {
		h.Logger.Info("Limit exceeded", zap.Int64("user", user.UserID))
		_, err := h.sendTextf(ctx, user, "Превышен лимит запросов в день (%d). Exceeded the limit of requests per day (%d).", h.limitPerDay, h.limitPerDay)

		if err != nil {
			h.Logger.Error("failed to send message", zap.Error(err))
		}

		return nil
	}

	if !cq && cqr == reasonPending {
		h.Logger.Info("Pending request", zap.Int64("user", user.UserID))
		_, err := h.sendText(ctx, user, "Предыдущий запрос еще выполняется, дождитесь завершения. Previous request is still in progress. Try again later.")

		if err != nil {
			h.Logger.Error("failed to send message", zap.Error(err))
		}

		return nil
	}

	if !cq && cqr == reasonNoUser {
		h.Logger.Error("No user data", zap.Int64("user", user.UserID))
		h.replyErrorf(ctx, user, nil, "Ошибка получения данных пользователя. Error getting user data.")
		return nil
	}

	h.incrPending(user.UserID)
	h.incrQueries(user.UserID)

	defer h.decrPending(user.UserID)

	var workingMessage *tg.Message
	var err error
	peer := h.inputUser(user)

	if workingMessage, err = h.sendText(ctx, user, "Работаю... Working..."); err != nil {
		h.Logger.Error("failed to send message", zap.Error(err))
		return nil
	}

	defer h.removeMessage(ctx, workingMessage)

	td, err := h.twitter.GetTwitterData(ctx, m.Message)

	if err != nil {
		h.Logger.Error("failed to get twitter data", zap.Error(err))
		h.replyError(ctx, user, err, "Ошибка получения данных из твиттера. Error getting data from twitter.")
		return errors.Wrap(err, "get twitter data")
	}

	if td.IsEmpty() {
		h.Logger.Info("empty tweet data")
		h.replyError(ctx, user, nil, "Не удалось получить данные из твиттера. Failed to get data from twitter.")
		return nil
	}

	h.Logger.Debug("twitter data", zap.Any("data", td))

	messageText := h.makeMessageText(td)

	if td.NoMedia() {
		_, err := h.sendText(ctx, user, messageText)
		if err != nil {
			h.Logger.Error("failed to send message", zap.Error(err))
		}
		return nil
	}

	downloads, err := h.downloader.DownloadTweetData(td, h.downloadFolder)

	if err != nil {
		h.Logger.Error("failed to download tweet data", zap.Error(err))
		h.replyError(ctx, user, err, "Ошибка cкачки. Error downloading.")
		return errors.Wrap(err, "download tweet data")
	}

	h.Logger.Info("Sending album", zap.Int("count", len(downloads)))

	uploads, err := h.uploadDownloads(ctx, downloads, messageText)

	if err != nil {
		h.Logger.Error("upload files", zap.Error(err))
		h.replyError(ctx, user, err, "Ошибка закачки в телеграм. Error uploading to telegram.")
		return errors.Wrap(err, "upload files")
	}

	sentMsgs, err := UnpackMultipleMessages(h.sender.To(peer).
		Album(ctx, uploads[0], uploads[1:]...))

	if err != nil {
		h.Logger.Error("send media group", zap.Error(err))
		h.replyError(ctx, user, err, "Ошибка отправки медигруппы в телеграм. Error sending media group.")
		return errors.Wrap(err, "send media group")
	}

	if h.forwardTo == 0 {
		return nil
	}

	sentMsgsIDs := make([]int, len(sentMsgs))

	for i, sentMsg := range sentMsgs {
		sentMsgsIDs[i] = sentMsg.ID
	}

	h.Logger.Info("Forwarding to channel", zap.Int64("channel", h.forwardTo))

	_, err = h.sender.To(h.inputChannelPeer()).
		ForwardIDs(h.inputUser(user), sentMsgsIDs[0], sentMsgsIDs[1:]...).
		Send(ctx)

	if err != nil {
		h.Logger.Error("forward", zap.Error(err))
		return errors.Wrap(err, "forward")
	}

	return nil
}
