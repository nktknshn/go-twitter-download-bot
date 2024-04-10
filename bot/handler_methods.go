package bot

import (
	"context"
	"fmt"
	"path"

	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/telegram/message/unpack"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
	"go.uber.org/zap"
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

func (h *Handler) OnNewMessage(ctx context.Context, entities tg.Entities, u *tg.UpdateNewMessage) error {
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

	// if user.UserID != h.AdminID {
	// not used in this bot
	// 	return nil
	// }

	if m.Message == "" {
		return nil
	}

	return h.OnNewMessageTextFromUser(ctx, entities, user, m)
}

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

func (h *Handler) inputUser(user *tg.PeerUser) tg.InputPeerClass {
	return &tg.InputPeerUser{
		UserID: user.UserID,
		// AccessHash: user.AccessHash,
	}
}

func (h *Handler) inputUserAdmin() tg.InputPeerClass {
	return &tg.InputPeerUser{
		UserID:     h.AdminID,
		AccessHash: 0,
	}
}
func (h *Handler) inputChannelPeer() tg.InputPeerClass {
	return &tg.InputPeerChannel{
		ChannelID:  h.UploadTo,
		AccessHash: h.UploadToAccessHash,
	}
}

func (h *Handler) InitSelfUsername(ctx context.Context) error {
	if h.selfUsername != "" {
		return nil
	}

	self, err := h.Api.UsersGetUsers(ctx, []tg.InputUserClass{
		&tg.InputUserSelf{},
	})

	if err != nil {
		return errors.Wrap(err, "get self")
	}

	user, ok := self[0].(*tg.User)
	if !ok {
		return errors.New("not a user")
	}

	h.selfUsername, ok = user.GetUsername()

	if !ok {
		return errors.New("no username")
	}

	return nil
}

func (h *Handler) InitChannelAccessHash(ctx context.Context) error {
	if h.UploadTo == 0 {
		return errors.New("channel id is not set")
	}

	if h.UploadToAccessHash != 0 {
		return nil
	}

	chatsClass, err := h.Api.ChannelsGetChannels(ctx, []tg.InputChannelClass{
		&tg.InputChannel{ChannelID: h.UploadTo},
	})

	if err != nil {
		return errors.Wrap(err, "get channel info")
	}

	chats := chatsClass.GetChats()

	if len(chats) == 0 {
		return errors.New("channel not found")
	}

	channel, ok := chats[0].(*tg.Channel)

	if !ok {
		return errors.New("not a channel")
	}

	h.UploadToAccessHash = channel.AccessHash

	return nil
}
