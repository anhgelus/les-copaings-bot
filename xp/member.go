package xp

import (
	"github.com/anhgelus/gokord"
	"github.com/anhgelus/gokord/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Copaing struct {
	gorm.Model
	DiscordID string `gorm:"not null"`
	XP        uint   `gorm:"default:0"`
	GuildID   string `gorm:"not null"`
}

var redisClient *redis.Client

func GetCopaing(discordID string, guildID string) *Copaing {
	c := Copaing{DiscordID: discordID, GuildID: guildID}
	return c.Load()
}

func (c *Copaing) Load() *Copaing {
	gokord.DB.Where("discord_id = ? and guild_id = ?", c.DiscordID, c.GuildID).FirstOrCreate(c)
	return c
}

func (c *Copaing) Save() {
	gokord.DB.Save(c)
}

func (c *Copaing) AddXP(s *discordgo.Session, m *discordgo.Member, xp uint, fn func(uint, uint)) {
	pastLevel := Level(c.XP)
	c.XP += xp
	c.Save()
	newLevel := Level(c.XP)
	if newLevel > pastLevel {
		fn(c.XP, newLevel)
		onNewLevel(s, m, newLevel)
	}
}

func getRedisClient() (*redis.Client, error) {
	if redisClient == nil {
		var err error
		redisClient, err = gokord.BaseCfg.Redis.Get()
		return redisClient, err
	}
	return redisClient, nil
}

func CloseRedisClient() {
	if redisClient == nil {
		return
	}
	err := redisClient.Close()
	if err != nil {
		utils.SendAlert("xp/member.go - Closing redis client", err.Error())
	}
}
