package config

import (
	"strconv"

	"github.com/anhgelus/gokord/cmd"
	"github.com/anhgelus/gokord/component"
	"github.com/anhgelus/gokord/logger"
	discordgo "github.com/nyttikord/gokord"
)

const (
	ModifyTimeReduce = "time_reduce"
	TimeReduceSet    = "time_reduce_set"
)

func HandleModifyPeriodicReduce(_ *discordgo.Session, _ *discordgo.InteractionCreate, _ discordgo.MessageComponentInteractionData, resp *cmd.ResponseBuilder) {
	err := resp.IsModal().
		SetCustomID(TimeReduceSet).
		SetComponents(component.New().ForModal().Add(component.NewActionRow().ForModal().Add(
			component.NewTextInput(TimeReduceSet, "Jours avant la réduction", discordgo.TextInputShort).
				SetMinLength(1).
				SetMaxLength(3),
		))).Send()
	if err != nil {
		logger.Alert("config/xp_reduce.go - Sending modal for periodic reduce", err.Error())
	}
}

func HandleTimeReduceSet(_ *discordgo.Session, i *discordgo.InteractionCreate, data discordgo.ModalSubmitInteractionData, resp *cmd.ResponseBuilder) {
	resp.IsEphemeral()
	v := data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
	days, err := strconv.Atoi(v)
	if err != nil {
		logger.Debug(err.Error())
		if err = resp.SetMessage("Nombres de jours invalides. Merci de mettre un entier.").Send(); err != nil {
			logger.Alert("config/xp_reduce.go - Sending bad input", err.Error())
		}
		return
	}
	if days < 30 {
		err = resp.SetMessage("Le nombre de jours est inférieur à 30.").Send()
		if err != nil {
			logger.Alert("config/xp_reduce.go - Days < 30 (fallback)", err.Error())
		}
		return
	}
	cfg := GetGuildConfig(i.GuildID)
	cfg.DaysXPRemains = uint(days)
	if err = cfg.Save(); err != nil {
		logger.Alert("config/channel.go - Saving days xp remains", err.Error())
		if err = resp.SetMessage("Erreur lors de la sauvegarde du salon").Send(); err != nil {
			logger.Alert("config/xp_reduce.go - Sending error while saving days xp remains", err.Error())
		}
		return
	}
	if err = resp.SetMessage("Modification sauvegardée.").Send(); err != nil {
		logger.Alert("config/xp_reduce.go - Sending days saved", err.Error())
	}
}
