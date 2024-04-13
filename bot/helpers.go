package bot

import (
	"github.com/go-faster/errors"
	"github.com/gotd/td/tg"
)

func UnpackMultipleMessages(update tg.UpdatesClass, err error) ([]*tg.Message, error) {
	if err != nil {
		return nil, err
	}

	var updates []tg.UpdateClass

	switch v := update.(type) {
	case *tg.UpdatesCombined:
		updates = v.GetUpdates()
	case *tg.Updates:
		updates = v.GetUpdates()
	default:
		return nil, errors.Errorf("unexpected type %T", update)
	}

	var messages []*tg.Message

	for _, update := range updates {
		switch v := update.(type) {
		case *tg.UpdateNewMessage:
			if m, ok := v.Message.(*tg.Message); ok {
				messages = append(messages, m)
			}
		case *tg.UpdateNewChannelMessage:
			if m, ok := v.Message.(*tg.Message); ok {
				messages = append(messages, m)
			}
		default:
		}
	}

	return messages, nil
}
