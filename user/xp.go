package user

import (
	"context"
	"errors"
	"fmt"
	"github.com/anhgelus/gokord"
	"github.com/anhgelus/gokord/utils"
	"github.com/anhgelus/les-copaings-bot/config"
	"github.com/anhgelus/les-copaings-bot/exp"
	"github.com/bwmarrin/discordgo"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"math"
	"strconv"
	"time"
)

func (c *Copaing) AddXP(s *discordgo.Session, m *discordgo.Member, xp uint, fn func(uint, uint)) {
	pastLevel := exp.Level(c.XP)
	old := c.XP
	c.XP += xp
	if err := c.Save(); err != nil {
		utils.SendAlert(
			"user/xp.go - Saving user",
			err.Error(),
			"exp",
			c.XP,
			"discord_id",
			c.DiscordID,
			"guild_id",
			c.GuildID,
		)
		c.XP = old
		return
	}
	newLevel := exp.Level(c.XP)
	if newLevel > pastLevel {
		fn(c.XP, newLevel)
		onNewLevel(s, m, newLevel)
	}
}

func (c *Copaing) SetLastEvent() {
	client, err := config.GetRedisClient()
	if err != nil {
		utils.SendAlert("user/xp.go - Getting redis client (set)", err.Error())
		return
	}
	t := time.Now().Unix()
	err = client.Set(context.Background(), c.GenKey(LastEvent), strconv.FormatInt(t, 10), 0).Err()
	if err != nil {
		utils.SendAlert("user/xp.go - Setting last event", err.Error(), "time", t, "base_key", c.GenKey(""))
		return
	}
	err = client.Set(context.Background(), c.GenKey(AlreadyRemoved), "0", 0).Err()
	if err != nil {
		utils.SendAlert(
			"user/xp.go - Setting already removed to 0",
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
	client, err := config.GetRedisClient()
	if err != nil {
		utils.SendAlert("user/xp.go - Getting redis client (get)", err.Error())
		return 0
	}
	res := client.Get(context.Background(), c.GenKey(LastEvent))
	if errors.Is(res.Err(), redis.Nil) {
		return 0
	} else if res.Err() != nil {
		utils.SendAlert("user/xp.go - Getting last event", res.Err().Error(), "base_key", c.GenKey(""))
		return 0
	}
	t := time.Now().Unix()
	last, err := strconv.Atoi(res.Val())
	if err != nil {
		utils.SendAlert(
			"user/xp.go - Converting time fetched into int (last event)",
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
	client, err := config.GetRedisClient()
	if err != nil {
		utils.SendAlert("user/xp.go - Getting redis client (set)", err.Error())
		return 0
	}
	exp := xp + c.XPAlreadyRemoved()
	err = client.Set(context.Background(), c.GenKey(AlreadyRemoved), exp, 0).Err()
	if err != nil {
		utils.SendAlert(
			"user/xp.go - Setting already removed",
			err.Error(),
			"exp already removed",
			exp,
			"base_key",
			c.GenKey(""),
		)
		return 0
	}
	return exp
}

func (c *Copaing) XPAlreadyRemoved() uint {
	client, err := config.GetRedisClient()
	if err != nil {
		utils.SendAlert("user/xp.go - Getting redis client (exp)", err.Error())
		return 0
	}
	res := client.Get(context.Background(), fmt.Sprintf("%s:%s", c.GenKey(""), AlreadyRemoved))
	if errors.Is(res.Err(), redis.Nil) {
		return 0
	} else if res.Err() != nil {
		utils.SendAlert("user/xp.go - Getting already removed", res.Err().Error(), "base_key", c.GenKey(""))
		return 0
	}
	xp, err := strconv.Atoi(res.Val())
	if err != nil {
		utils.SendAlert(
			"user/xp.go - Converting time fetched into int (already removed)",
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
			"user/xp.go - Assertion exp >= 0",
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

func (c *Copaing) AfterDelete(db *gorm.DB) error {
	id := c.ID
	dID := c.DiscordID
	gID := c.GuildID
	k := c.GuildID + ":" + c.DiscordID
	ch := utils.NewTimer(48*time.Hour, func(stop chan<- interface{}) {
		if err := db.Unscoped().Where("id = ?", id).Delete(c).Error; err != nil {
			utils.SendAlert(
				"user/xp.go - Removing user from database", err.Error(),
				"discord_id", dID,
				"guild_id", gID,
			)
		}
		stop <- true
		leftCopaingsMap[k] = nil
	})
	leftCopaingsMap[k] = &leftCopaing{id, ch}
	return nil
}
