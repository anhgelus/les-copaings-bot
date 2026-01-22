package config

import (
	"context"
	"fmt"
	"strconv"

	"github.com/nyttikord/gokord/bot"
	"github.com/nyttikord/gokord/component"
	"github.com/nyttikord/gokord/interaction"
)

const (
	ModifyTimeReduce = "time_reduce"
	TimeReduceSet    = "time_reduce_set"
)

func HandleModifyPeriodicReduceCommand(ctx context.Context, dg bot.Session, i *interaction.MessageComponent) {
	cfg := GetGuildConfig(i.GuildID)
	resp := interaction.NewModalResponse().
		CustomID(TimeReduceSet).
		Title("Modifier la durée de l'expérience").
		AddComponent(&component.TextDisplay{
			Content: "Seul l'expérience gagnée sur cette période sera comptabilisée dans le niveau par défaut",
		}).
		AddComponent(&component.Label{
			Label: "Durée en jours",
			Component: &component.TextInput{
				CustomID:    TimeReduceSet,
				MinLength:   1,
				MaxLength:   3,
				Style:       component.TextInputShort,
				Placeholder: "Durée en jours",
				Value:       fmt.Sprintf("%d", cfg.DaysXPRemains),
			},
		}).
		Response()
	err := dg.InteractionAPI().Respond(i.Interaction, resp).Do(ctx)
	if err != nil {
		bot.Logger(ctx).Error("sending xp reduce modal", "error", err)
	}
}

func HandleTimeReduceSet(ctx context.Context, dg bot.Session, i *interaction.ModalSubmit) bool {
	v := i.Data.Components[0].(*component.Label).Component.(*component.TextInput).Value
	days, err := strconv.Atoi(v)
	if err != nil {
		resp := interaction.NewMessageResponse().
			IsEphemeral().
			Message(fmt.Sprintf("La valeur indiquée, `%s`, c'est pas un entier.", v)).
			Response()
		err = dg.InteractionAPI().Respond(i.Interaction, resp).Do(ctx)
		if err != nil {
			bot.Logger(ctx).Error("sending bad input message", "error", err)
		}
		return false
	}
	if days < 30 {
		resp := interaction.NewMessageResponse().
			IsEphemeral().
			Message("Le nombre de jours doit être suppérieur à 30.").
			Response()
		err = dg.InteractionAPI().Respond(i.Interaction, resp).Do(ctx)
		if err != nil {
			bot.Logger(ctx).Error("sending less than 30 days message", "error", err)
		}
		return false
	}
	cfg := GetGuildConfig(i.GuildID)
	cfg.DaysXPRemains = uint(days)
	err = cfg.Save()
	if err != nil {
		bot.Logger(ctx).Error("saving DaysXPRemains configuration", "error", err)
		return false
	}
	return true
}
