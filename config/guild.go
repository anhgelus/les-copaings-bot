package config

import (
	"strings"

	"github.com/anhgelus/gokord"
)

type GuildConfig struct {
	ID               uint   `gorm:"primarykey"`
	GuildID          string `gorm:"not null;unique"`
	XpRoles          []XpRole
	DisabledChannels string
	FallbackChannel  string
	DaysXPRemains    uint `gorm:"default:90"` // 30 * 3 = 90 (three months)
}

func GetGuildConfig(guildID string) *GuildConfig {
	cfg := GuildConfig{GuildID: guildID}
	if err := cfg.Load(); err != nil {
		panic(err)
	}
	return &cfg
}

func (cfg *GuildConfig) Load() error {
	return gokord.DB.Where("guild_id = ?", cfg.GuildID).Preload("XpRoles").FirstOrCreate(cfg).Error
}

func (cfg *GuildConfig) Save() error {
	return gokord.DB.Save(cfg).Error
}

func (cfg *GuildConfig) IsDisabled(channelID string) bool {
	return strings.Contains(cfg.DisabledChannels, channelID)
}

func (cfg *GuildConfig) FindXpRole(roleID string) (int, *XpRole) {
	for i, r := range cfg.XpRoles {
		if r.RoleID == roleID {
			return i, &r
		}
	}
	return 0, nil
}
