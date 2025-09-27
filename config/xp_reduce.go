package config

import (
	"fmt"
	"strconv"

	"github.com/anhgelus/gokord/cmd"
	"github.com/nyttikord/gokord/bot"
	"github.com/nyttikord/gokord/component"
	"github.com/nyttikord/gokord/discord/types"
	"github.com/nyttikord/gokord/event"
	"github.com/nyttikord/gokord/interaction"
)

const (
	ModifyTimeReduce = "time_reduce"
	TimeReduceSet    = "time_reduce_set"
)

func HandleModifyPeriodicReduceCommand(s bot.Session, i *event.InteractionCreate, _ *interaction.MessageComponentData, _ *cmd.ResponseBuilder) {
	cfg := GetGuildConfig(i.GuildID)
	response := interaction.Response{
		Type: types.InteractionResponseModal,
		Data: &interaction.ResponseData{
			CustomID: TimeReduceSet,
			Title:    "Modifier la durée de l'expérience",
			Components: []component.Component{
				// TODO: When gokord supports it, enable this description again
				// &component.TextDisplay{
				// 	Content: "Seul l'expérience gagnée sur cette période sera comptabilisée dans le niveau par défaut",
				// },
				&component.Label{
					Label: "Durée en jours",
					Component: &component.TextInput{
						CustomID:    TimeReduceSet,
						MinLength:   1,
						MaxLength:   3,
						Style:       component.TextInputShort,
						Placeholder: "Durée en jours",
						Value:       fmt.Sprintf("%d", cfg.DaysXPRemains),
					},
				},
			},
		},
	}
	err := s.InteractionAPI().Respond(i.Interaction, &response)
	if err != nil {
		s.Logger().Error("sending xp reduce modal", "error", err)
	}
}

func HandleTimeReduceSet(s bot.Session, i *event.InteractionCreate, data *interaction.ModalSubmitData, resp *cmd.ResponseBuilder) bool {
	v := data.Components[0].(*component.Label).Component.(*component.TextInput).Value
	days, err := strconv.Atoi(v)
	if err != nil {
		err = resp.IsEphemeral().SetMessage(fmt.Sprintf("La valeur indiquée, `%s`, c'est pas un entier.", v)).Send()
		if err != nil {
			s.Logger().Error("sending bad input message", "error", err)
		}
		return false
	}
	if days < 30 {
		err = resp.IsEphemeral().SetMessage("Le nombre de jours doit être suppérieur à 30.").Send()
		if err != nil {
			s.Logger().Error("sending less than 30 days message", "error", err)
		}
		return false
	}
	cfg := GetGuildConfig(i.GuildID)
	cfg.DaysXPRemains = uint(days)
	err = cfg.Save()
	if err != nil {
		s.Logger().Error("saving DaysXPRemains configuration", "error", err)
		return false
	}
	return true
}
