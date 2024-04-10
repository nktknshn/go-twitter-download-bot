package bot

import (
	"context"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/unpack"
	"github.com/gotd/td/tg"
	"github.com/nktknshn/go-twitter-fun/twitter"
	"go.uber.org/zap"
)

type Handler struct {
	Logger *zap.Logger

	Dispatcher tg.UpdateDispatcher
	Api        *tg.Client
	Sender     *message.Sender

	DebugTelegram      bool
	AdminID            int64
	UploadTo           int64
	UploadToAccessHash int64
	DownloadFolder     string

	twitter    *twitter.Twitter
	downloader *Downloader

	selfUsername string
}

func (h *Handler) botName() string {
	return h.selfUsername
}

func (h *Handler) Init(client *telegram.Client) {
	h.Api = tg.NewClient(client)
	h.Sender = message.NewSender(h.Api)
	h.twitter = twitter.NewTwitter()
	h.downloader = NewDownloader()
	h.Dispatcher.OnNewMessage(h.OnNewMessage)
}

func (h *Handler) Handle(ctx context.Context, u tg.UpdatesClass) error {
	if h.DebugTelegram {
		h.Logger.Debug("update", zap.Any("update", u))
	}
	return h.Dispatcher.Handle(ctx, u)
}

func (h *Handler) OnStart(ctx context.Context, entities tg.Entities, user *tg.PeerUser, m *tg.Message) error {
	msg := "Отправь ссылку на пост в твиттер и я скачаю фото или видео.\nSend me a link to a tweet and I will download the photo or video."
	if _, err := h.ReplyText(ctx, user, msg); err != nil {
		h.Logger.Error("failed to send message", zap.Error(err))
	}
	return nil
}

func (h *Handler) ReplyText(ctx context.Context, user *tg.PeerUser, text string) (*tg.Message, error) {
	return unpack.Message(h.Sender.To(h.inputUser(user)).Text(ctx, text))
}

func (h *Handler) ReplyError(ctx context.Context, user *tg.PeerUser, err error, msg string) {
	_, _ = h.ReplyText(ctx, user, "Ошибка. Error. "+msg)
}

func (h *Handler) OnNewMessageTextFromUser(ctx context.Context, entities tg.Entities, user *tg.PeerUser, m *tg.Message) error {
	if m.Message == "/start" {
		return h.OnStart(ctx, entities, user, m)
	}

	if !twitter.IsValidTwitterURL(m.Message) {
		return nil
	}

	return h.OnTwitterURLFromUser(ctx, entities, user, m)
}