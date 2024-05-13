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

	if h.forwardTo == 0 {
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

	self, err := h.api.UsersGetUsers(ctx, []tg.InputUserClass{&tg.InputUserSelf{}})

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
	if h.forwardTo == 0 {
		return errors.New("channel id is not set")
	}

	if h.uploadToAccessHash != 0 {
		return nil
	}

	chatsClass, err := h.api.ChannelsGetChannels(ctx, []tg.InputChannelClass{
		&tg.InputChannel{ChannelID: h.forwardTo},
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

	h.uploadToAccessHash = channel.AccessHash

	return nil
}
