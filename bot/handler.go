package bot

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/unpack"
	"github.com/gotd/td/tg"
	"github.com/nktknshn/go-twitter-download-bot/twitter"
	"go.uber.org/zap"
)

type Handler struct {
	Logger *zap.Logger

	dispatcher tg.UpdateDispatcher
	api        *tg.Client
	sender     *message.Sender

	debugTelegram     bool
	adminID           int64
	restrictToAdminID bool

	forwardTo          int64
	uploadToAccessHash int64
	downloadFolder     string

	twitter    *twitter.Twitter
	downloader *Downloader

	selfUsername string

	IncludeText    bool
	IncludeURL     bool
	IncludeBotName bool

	limitPerDay  int
	limitPending int

	usersMap     map[int64]*UserData
	usersMapLock sync.RWMutex

	nowFunc func() time.Time
}

func (h *Handler) botName() string {
	return h.selfUsername
}

func (h *Handler) Init(client *telegram.Client) error {

	if h.nowFunc == nil {
		h.nowFunc = time.Now
	}

	h.usersMap = make(map[int64]*UserData)
	h.usersMapLock = sync.RWMutex{}

	h.api = tg.NewClient(client)
	h.sender = message.NewSender(h.api)
	h.twitter = twitter.NewTwitter()
	h.downloader = NewDownloader()
	h.dispatcher.OnNewMessage(h.onNewMessage)

	return nil
}

func (h *Handler) Handle(ctx context.Context, u tg.UpdatesClass) error {
	if h.debugTelegram {
		h.Logger.Debug("update", zap.Any("update", u))
	}
	return h.dispatcher.Handle(ctx, u)
}

func (h *Handler) onNewMessage(ctx context.Context, entities tg.Entities, u *tg.UpdateNewMessage) error {
	m, ok := u.Message.(*tg.Message)

	if !ok {
		return nil
	}

	if m.Out {
		return nil
	}

	user, ok := m.PeerID.(*tg.PeerUser)

	if !ok {
		return nil
	}

	if h.adminRestricted() && !h.isAdmin(user.UserID) {
		return nil
	}

	if m.Message == "" {
		return nil
	}

	return h.onNewMessageTextFromUser(ctx, entities, user, m)
}

func (h *Handler) onNewMessageTextFromUser(ctx context.Context, entities tg.Entities, user *tg.PeerUser, m *tg.Message) error {

	h.initUser(user.UserID)

	if m.Message == "/start" {
		return h.onStart(ctx, entities, user, m)
	}

	if !twitter.IsValidTwitterURL(m.Message) {
		return nil
	}

	return h.onTwitterURLFromUser(ctx, entities, user, m)
}

func (h *Handler) onStart(ctx context.Context, entities tg.Entities, user *tg.PeerUser, m *tg.Message) error {
	msg := "Отправь ссылку на пост в твиттер и я скачаю фото или видео.\nSend me a link to a tweet and I will download the photo or video."
	if _, err := h.sendText(ctx, user, msg); err != nil {
		h.Logger.Error("failed to send message", zap.Error(err))
	}
	return nil
}
func (h *Handler) sendTextf(ctx context.Context, user *tg.PeerUser, format string, args ...interface{}) (*tg.Message, error) {
	return h.sendText(ctx, user, fmt.Sprintf(format, args...))
}

func (h *Handler) sendText(ctx context.Context, user *tg.PeerUser, text string) (*tg.Message, error) {
	return unpack.Message(h.sender.To(h.inputUser(user)).Text(ctx, text))
}

func (h *Handler) replyErrorf(ctx context.Context, user *tg.PeerUser, err error, format string, args ...interface{}) {
	h.replyError(ctx, user, err, fmt.Sprintf(format, args...))
}

func (h *Handler) replyError(ctx context.Context, user *tg.PeerUser, err error, msg string) {
	_, _ = h.sendText(ctx, user, msg+" Попробуйте отправить ещё раз. Try again.")
}
