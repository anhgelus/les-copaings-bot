package config

import (
	"fmt"

	"github.com/anhgelus/gokord"
	"github.com/anhgelus/gokord/cmd"
	discordgo "github.com/nyttikord/gokord"
	"github.com/nyttikord/gokord/interaction"
)

type XpRole struct {
	ID            uint `gorm:"primarykey"`
	XP            uint
	RoleID        string
	GuildConfigID uint
}

const (
	ModifyXpRole    = "xp_role"
	XpRoleAdd       = "xp_role_add"
	XpRoleAddLevel  = "xp_role_add_level"
	XpRoleAddRole   = "xp_role_add_role"
	XpRoleDel       = "xp_role_del"
	XpRoleDelRole   = "xp_role_del_role"
	XpRoleEdit      = "xp_role_edit"
	XpRoleEditLevel = "xp_role_edit_level"
	XpRoleEditRole  = "xp_role_edit_role"
)

var (
	configModifyMap = map[string]uint{}
)

func HandleModifyXpRole(_ *discordgo.Session, _ *discordgo.InteractionCreate, _ interaction.MessageComponentData, resp *cmd.ResponseBuilder) {
	//err := resp.IsEphemeral().
	//	SetMessage("Action à réaliser").
	//	SetComponents(component.New().Add(component.NewActionRow().
	//		Add(component.NewButton(XpRoleAdd, discordgo.PrimaryButton).
	//			SetLabel("Ajouter").
	//			SetEmoji(&discordgo.ComponentEmoji{Name: "⬆️"}),
	//		).
	//		Add(component.NewButton(XpRoleEdit, discordgo.SecondaryButton).
	//			SetLabel("Modifier").
	//			SetEmoji(&discordgo.ComponentEmoji{Name: "📝"}),
	//		).
	//		Add(component.NewButton(XpRoleDel, discordgo.DangerButton).
	//			SetLabel("Supprimer").
	//			SetEmoji(&discordgo.ComponentEmoji{Name: "❌"}),
	//		),
	//	)).Send()
	//if err != nil {
	//	logger.Alert("config/xp_reduce.go - Sending config", err.Error())
	//}
}

func HandleXpRoleAddEdit(_ *discordgo.Session, _ *discordgo.InteractionCreate, data interaction.MessageComponentData, resp *cmd.ResponseBuilder) {
	//cID := XpRoleAddLevel
	//if data.CustomID == XpRoleEdit {
	//	cID = XpRoleEditLevel
	//}
	//err := resp.IsModal().
	//	SetTitle("Role").
	//	SetCustomID(cID).
	//	SetComponents(component.New().ForModal().Add(component.NewActionRow().ForModal().Add(
	//		component.NewTextInput(cID, "Niveau", discordgo.TextInputShort).
	//			SetPlaceholder("5").
	//			IsRequired().
	//			SetMinLength(0).
	//			SetMaxLength(5),
	//	))).
	//	Send()
	//if err != nil {
	//	logger.Alert("config/xp_reduce.go - Sending modal to add/edit", err.Error())
	//}
}

func HandleXpRoleAddRole(s *discordgo.Session, i *discordgo.InteractionCreate, data interaction.MessageComponentData, resp *cmd.ResponseBuilder) {
	resp.IsEphemeral()
	cfg := GetGuildConfig(i.GuildID)
	roleId := data.Values[0]
	for _, r := range cfg.XpRoles {
		if r.RoleID == roleId {
			err := resp.SetMessage("Le rôle est déjà présent dans la config").Send()
			if err != nil {
				s.LogError(err, "sending role already in config")
			}
			return
		}
	}
	cfg.XpRoles = append(cfg.XpRoles, XpRole{
		XP:     configModifyMap[getKeyConfigRole(i)],
		RoleID: roleId,
	})
	err := cfg.Save()
	if err != nil {
		s.LogError(err, "saving config for role %s in %s", roleId, i.GuildID)
		return
	}
	if err = resp.SetMessage("Rôle ajouté.").Send(); err != nil {
		s.LogError(err, "Sending role saved")
	}
}

func HandleXpRoleEditRole(s *discordgo.Session, i *discordgo.InteractionCreate, data interaction.MessageComponentData, resp *cmd.ResponseBuilder) {
	resp.IsEphemeral()
	cfg := GetGuildConfig(i.GuildID)
	roleId := data.Values[0]
	_, r := cfg.FindXpRole(roleId)
	if r == nil {
		err := resp.SetMessage("Le rôle n'a pas été trouvé dans la config.").Send()
		if err != nil {
			s.LogError(err, "role not found")
		}
		return
	}
	r.XP = configModifyMap[getKeyConfigRole(i)]
	err := gokord.DB.Save(r).Error
	if err != nil {
		s.LogError(err, "saving config for role %s in %s", roleId, i.GuildID)
		return
	}
	if err = resp.SetMessage("Rôle modifié.").Send(); err != nil {
		s.LogError(err, "sending role saved")
	}
}

func HandleXpRoleDel(_ *discordgo.Session, _ *discordgo.InteractionCreate, _ interaction.MessageComponentData, resp *cmd.ResponseBuilder) {
	//err := resp.IsEphemeral().
	//	SetMessage("Rôle à supprimer").
	//	SetComponents(component.New().Add(component.NewActionRow().Add(component.NewRoleSelect(XpRoleDelRole)))).
	//	Send()
	//if err != nil {
	//	logger.Alert("config/xp_reduce.go - Sending response to del", err.Error())
	//}
}

func HandleXpRoleDelRole(s *discordgo.Session, i *discordgo.InteractionCreate, data interaction.MessageComponentData, resp *cmd.ResponseBuilder) {
	resp.IsEphemeral()
	cfg := GetGuildConfig(i.GuildID)
	roleId := data.Values[0]
	_, r := cfg.FindXpRole(roleId)
	if r == nil {
		err := resp.SetMessage("Le rôle n'a pas été trouvé dans la config.").Send()
		if err != nil {
			s.LogError(err, "sending role not found")
		}
		return
	}
	err := gokord.DB.Delete(r).Error
	if err != nil {
		s.LogError(err, "saving config for role %s in %s", roleId, i.GuildID)
		return
	}
	if err = resp.SetMessage("Rôle supprimé.").Send(); err != nil {
		s.LogError(err, "sending role deleted")
	}
}

func HandleXpRoleLevel(_ *discordgo.Session, i *discordgo.InteractionCreate, data interaction.ModalSubmitData, resp *cmd.ResponseBuilder) {
	//resp.IsEphemeral()
	//input := data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput)
	//
	//k := getKeyConfigRole(i)
	//in, err := strconv.Atoi(input.Value)
	//if err != nil || in < 0 {
	//	if err = resp.
	//		SetMessage("Impossible de lire le nombre. Il doit s'agit d'un nombre entier positif.").
	//		Send(); err != nil {
	//		logger.Alert("command/config.go - Sending bad number", err.Error())
	//	}
	//	return
	//}
	//configModifyMap[k] = exp.LevelXP(uint(in))
	//go func(i *discordgo.InteractionCreate, k string) {
	//	time.Sleep(5 * time.Minute)
	//	delete(configModifyMap, k)
	//}(i, k)
	//
	//cID := XpRoleAddRole
	//resp.SetMessage("Rôle à ajouter")
	//if data.CustomID == XpRoleEditLevel {
	//	cID = XpRoleEditRole
	//	resp.SetMessage("Rôle à modifier")
	//}
	//
	//err = resp.
	//	SetComponents(component.New().Add(component.NewActionRow().Add(component.NewRoleSelect(cID)))).
	//	Send()
	//if err != nil {
	//	logger.Alert("config/xp_reduce.go - Sending response to add/edit", err.Error())
	//}
}

func getKeyConfigRole(i *discordgo.InteractionCreate) string {
	return fmt.Sprintf("r:%s:%s", i.GuildID, i.Member.User.ID)
}
