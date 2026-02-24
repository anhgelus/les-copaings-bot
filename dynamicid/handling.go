package dynamicid

import (
	"context"
	"strings"

	"github.com/nyttikord/gokord/bot"
	"github.com/nyttikord/gokord/discord/types"
	"github.com/nyttikord/gokord/interaction"
	"github.com/nyttikord/gokord/interaction/interactionhandler"
)

func HandleDynamicMessageComponent[T any](
	m *interactionhandler.Manager,
	handler func(context.Context, bot.Session, *interaction.MessageComponent, T),
	base string,
) {
	m.HandleRaw(func(ctx context.Context, dg bot.Session, i *interaction.Interaction) {
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
	m *interactionhandler.Manager,
	handler func(context.Context, bot.Session, *interaction.ModalSubmit, T),
	base string,

) {
	m.HandleRaw(func(ctx context.Context, dg bot.Session, i *interaction.Interaction) {
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
