package bot

import (
	"context"

	"github.com/gotd/td/tg"
	"go.uber.org/zap"
)

func (h *Handler) removeMessage(ctx context.Context, m *tg.Message) {
	h.Logger.Info("Removing message", zap.Any("message", m.ID))

	if _, err := h.api.MessagesDeleteMessages(ctx, &tg.MessagesDeleteMessagesRequest{
		ID:     []int{m.ID},
		Revoke: true,
	}); err != nil {
		h.Logger.Error("failed to remove working message", zap.Error(err))
	}
}

func (h *Handler) inputUser(user *tg.PeerUser) tg.InputPeerClass {
	return &tg.InputPeerUser{
		UserID: user.UserID,
		// AccessHash: user.AccessHash,
	}
}

func (h *Handler) inputUserAdmin() tg.InputPeerClass {
	return &tg.InputPeerUser{
		UserID:     h.adminID,
		AccessHash: 0,
	}
}
func (h *Handler) inputChannelPeer() tg.InputPeerClass {
	return &tg.InputPeerChannel{
		ChannelID:  h.forwardTo,
		AccessHash: h.uploadToAccessHash,
	}
}
