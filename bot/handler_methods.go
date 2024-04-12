package bot

import (
	"github.com/gotd/td/tg"
)

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
