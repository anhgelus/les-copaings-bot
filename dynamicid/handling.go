package dynamicid

import (
	"strings"

	"github.com/anhgelus/gokord"
	"github.com/anhgelus/gokord/cmd"
	"github.com/nyttikord/gokord/bot"
	"github.com/nyttikord/gokord/discord/types"
	"github.com/nyttikord/gokord/event"
	"github.com/nyttikord/gokord/interaction"
)

func HandleDynamicMessageComponent[DynamicData any](
	b *gokord.Bot,
	handler func(
		bot.Session,
		*event.InteractionCreate,
		*interaction.MessageComponentData,
		*DynamicData, *cmd.ResponseBuilder,
	),
	base string,
) {
	b.AddHandler(func(s bot.Session, i *event.InteractionCreate) {
		if i.Type != types.InteractionMessageComponent {
			return
		}
		data := i.MessageComponentData()
		if !strings.HasPrefix(data.CustomID, base+";") {
			return
		}
		dynamicID := data.CustomID[len(base)+1:]
		dynamicData := new(DynamicData)
		err := UnmarshallCSV(dynamicID, dynamicData)
		if err != nil {
			s.Logger().Error("Unable to parse CustomID", "error", err, "CustomID", data.CustomID, "base", base)
			return
		}
		handler(s, i, data, dynamicData, cmd.NewResponseBuilder(s, i))
	})
}

func HandleDynamicModalComponent[DynamicData any](
	b *gokord.Bot,
	handler func(
		bot.Session,
		*event.InteractionCreate,
		*interaction.ModalSubmitData,
		*DynamicData,
		*cmd.ResponseBuilder,
	),
	base string,
) {
	b.AddHandler(func(s bot.Session, i *event.InteractionCreate) {
		if i.Type != types.InteractionModalSubmit {
			return
		}
		data := i.ModalSubmitData()
		if strings.HasPrefix(data.CustomID, base+";") {
			dynamicID := data.CustomID[len(base)+1:]
			dynamicData := new(DynamicData)
			err := UnmarshallCSV(dynamicID, dynamicData)
			if err != nil {
				s.Logger().Error("Unable to parse CustomID", "error", err, "CustomID", data.CustomID, "base", base)
				return
			}
			handler(s, i, data, dynamicData, cmd.NewResponseBuilder(s, i))
		}
	})
}

func FormatCustomID(base string, dynamicData any) string {
	return base + ";" + MarshallCSV(dynamicData)
}
