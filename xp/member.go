package xp

import (
	"context"
	"errors"
	"fmt"
	"github.com/anhgelus/gokord"
	"github.com/anhgelus/gokord/utils"
	"github.com/bwmarrin/discordgo"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"math"
	"strconv"
	"time"
)

type Copaing struct {
	gorm.Model
	DiscordID string `gorm:"not null"`
	XP        uint   `gorm:"default:0"`
	GuildID   string `gorm:"not null"`
}

var redisClient *redis.Client

const (
	LastEvent      = "last_event"
	AlreadyRemoved = "already_removed"
)

func GetCopaing(discordID string, guildID string) *Copaing {
	c := Copaing{DiscordID: discordID, GuildID: guildID}
	if err := c.Load(); err != nil {
		utils.SendAlert(
			"xp/member.go - Loading copaing",
			err.Error(),
			"discord_id",
			discordID,
			"guild_id",
			guildID,
		)
		return nil
	}
	return &c
}

func (c *Copaing) Load() error {
	return gokord.DB.Where("discord_id = ? and guild_id = ?", c.DiscordID, c.GuildID).FirstOrCreate(c).Error
}

func (c *Copaing) Save() error {
	return gokord.DB.Save(c).Error
}

func (c *Copaing) GenKey(key string) string {
	return fmt.Sprintf("%s:%s:%s", c.GuildID, c.DiscordID, key)
}

func (c *Copaing) AddXP(s *discordgo.Session, m *discordgo.Member, xp uint, fn func(uint, uint)) {
	pastLevel := Level(c.XP)
	old := c.XP
	c.XP += xp
	if err := c.Save(); err != nil {
		utils.SendAlert(
			"xp/level.go - Saving copaing",
			err.Error(),
			"xp",
			c.XP,
			"discord_id",
			c.DiscordID,
			"guild_id",
			c.GuildID,
		)
		c.XP = old
		return
	}
	newLevel := Level(c.XP)
	if newLevel > pastLevel {
		fn(c.XP, newLevel)
		onNewLevel(s, m, newLevel)
	}
}

func (c *Copaing) SetLastEvent() {
	client, err := getRedisClient()
	if err != nil {
		utils.SendAlert("xp/member.go - Getting redis client (set)", err.Error())
		return
	}
	t := time.Now().Unix()
	err = client.Set(context.Background(), c.GenKey(LastEvent), strconv.FormatInt(t, 10), 0).Err()
	if err != nil {
		utils.SendAlert("xp/member.go - Setting last event", err.Error(), "time", t, "base_key", c.GenKey(""))
		return
	}
	err = client.Set(context.Background(), c.GenKey(AlreadyRemoved), "0", 0).Err()
	if err != nil {
		utils.SendAlert(
			"xp/member.go - Setting already removed to 0",
			err.Error(),
			"time",
			t,
			"base_key",
			c.GenKey(""),
		)
		return
	}
}

func (c *Copaing) HourSinceLastEvent() uint {
	client, err := getRedisClient()
	if err != nil {
		utils.SendAlert("xp/member.go - Getting redis client (get)", err.Error())
		return 0
	}
	res := client.Get(context.Background(), c.GenKey(LastEvent))
	if errors.Is(res.Err(), redis.Nil) {
		return 0
	} else if res.Err() != nil {
		utils.SendAlert("xp/member.go - Getting last event", res.Err().Error(), "base_key", c.GenKey(""))
		return 0
	}
	t := time.Now().Unix()
	last, err := strconv.Atoi(res.Val())
	if err != nil {
		utils.SendAlert(
			"xp/member.go - Converting time fetched into int (last event)",
			err.Error(),
			"base_key",
			c.GenKey(""),
			"val",
			res.Val(),
		)
		return 0
	}
	if gokord.Debug {
		return uint(math.Floor(float64(t-int64(last)) / 60)) // not hours of unix, is minutes of unix
	}
	return utils.HoursOfUnix(t - int64(last))
}

func (c *Copaing) AddXPAlreadyRemoved(xp uint) uint {
	client, err := getRedisClient()
	if err != nil {
		utils.SendAlert("xp/member.go - Getting redis client (set)", err.Error())
		return 0
	}
	exp := xp + c.XPAlreadyRemoved()
	err = client.Set(context.Background(), c.GenKey(AlreadyRemoved), exp, 0).Err()
	if err != nil {
		utils.SendAlert(
			"xp/member.go - Setting already removed",
			err.Error(),
			"xp already removed",
			exp,
			"base_key",
			c.GenKey(""),
		)
		return 0
	}
	return exp
}

func (c *Copaing) XPAlreadyRemoved() uint {
	client, err := getRedisClient()
	if err != nil {
		utils.SendAlert("xp/member.go - Getting redis client (xp)", err.Error())
		return 0
	}
	res := client.Get(context.Background(), fmt.Sprintf("%s:%s", c.GenKey(""), AlreadyRemoved))
	if errors.Is(res.Err(), redis.Nil) {
		return 0
	} else if res.Err() != nil {
		utils.SendAlert("xp/member.go - Getting already removed", res.Err().Error(), "base_key", c.GenKey(""))
		return 0
	}
	xp, err := strconv.Atoi(res.Val())
	if err != nil {
		utils.SendAlert(
			"xp/member.go - Converting time fetched into int (already removed)",
			err.Error(),
			"base_key",
			c.GenKey(""),
			"val",
			res.Val(),
		)
		return 0
	}
	if xp < 0 {
		utils.SendAlert(
			"xp/member.go - Assertion xp >= 0",
			"xp is negative",
			"base_key",
			c.GenKey(""),
			"xp",
			xp,
		)
		return 0
	}
	return uint(xp)
}

func (c *Copaing) Reset() {
	gokord.DB.Where("guild_id = ? AND discord_id = ?", c.GuildID, c.DiscordID).Delete(c)
}

func getRedisClient() (*redis.Client, error) {
	if redisClient == nil {
		var err error
		redisClient, err = gokord.BaseCfg.GetRedisCredentials().Connect()
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
