package bot

import (
	"context"

	"github.com/go-faster/errors"
	"github.com/gotd/td/tg"
)

func (h *Handler) OnConnected(ctx context.Context) error {
	if err := h.initSelfUsername(ctx); err != nil {
		return errors.Wrap(err, "init self username")
	}

	if h.ForwardTo == 0 {
		return nil
	}

	if err := h.initChannelAccessHash(ctx); err != nil {
		return errors.Wrap(err, "init channel access hash")
	}

	return nil
}

func (h *Handler) initSelfUsername(ctx context.Context) error {
	if h.selfUsername != "" {
		return nil
	}

	self, err := h.Api.UsersGetUsers(ctx, []tg.InputUserClass{&tg.InputUserSelf{}})

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

func (h *Handler) initChannelAccessHash(ctx context.Context) error {
	if h.ForwardTo == 0 {
		return errors.New("channel id is not set")
	}

	if h.UploadToAccessHash != 0 {
		return nil
	}

	chatsClass, err := h.Api.ChannelsGetChannels(ctx, []tg.InputChannelClass{
		&tg.InputChannel{ChannelID: h.ForwardTo},
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
