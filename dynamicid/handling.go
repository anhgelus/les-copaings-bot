package dynamicid

import (
	"context"
	"strings"

	"github.com/nyttikord/gokord/bot"
	"github.com/nyttikord/gokord/discord/types"
	"github.com/nyttikord/gokord/event"
	"github.com/nyttikord/gokord/interaction"
)

func HandleDynamicMessageComponent[T any](
	events bot.EventManager,
	handler func(context.Context, bot.Session, *interaction.MessageComponent, T),
	base string,
) {
	events.AddHandler(func(ctx context.Context, dg bot.Session, i *event.InteractionCreate) {
		if i.Type != types.InteractionMessageComponent {
			return
		}
		msg := i.MessageComponent()
		cid := msg.Data.CustomID
		if !strings.HasPrefix(cid, base+";") {
			return
		}
		dynamicID := cid[len(base)+1:]
		var dynamic T
		err := UnmarshallCSV(dynamicID, dynamic)
		if err != nil {
			bot.Logger(ctx).Error("Unable to parse CustomID", "error", err, "CustomID", cid, "base", base)
			return
		}
		handler(ctx, dg, msg, dynamic)
	})
}

func HandleDynamicModalComponent[T any](
	events bot.EventManager,
	handler func(context.Context, bot.Session, *interaction.ModalSubmit, T),
	base string,

) {
	events.AddHandler(func(ctx context.Context, dg bot.Session, i *event.InteractionCreate) {
		if i.Type != types.InteractionModalSubmit {
			return
		}
		modal := i.ModalSubmit()
		cid := modal.Data.CustomID
		if !strings.HasPrefix(cid, base+";") {
			return
		}
		dynamicID := cid[len(base)+1:]
		var dynamic T
		err := UnmarshallCSV(dynamicID, dynamic)
		if err != nil {
			bot.Logger(ctx).Error("Unable to parse CustomID", "error", err, "CustomID", cid, "base", base)
			return
		}
		handler(ctx, dg, modal, dynamic)
	})
}

func FormatCustomID(base string, dynamicData any) string {
	return base + ";" + MarshallCSV(dynamicData)
}
