package user

import (
	"sync"

	"github.com/nyttikord/gokord/state"
)

type CopaingCached struct {
	ID        uint   `gorm:"primarykey"`
	DiscordID string `gorm:"not null"`
	GuildID   string `gorm:"not null"`
	XPs       uint
	XPToAdd   uint
}

const KeyCopaingCachedPrefix = "cc:"

func KeyCopaingCached(c *Copaing) state.Key {
	return KeyCopaingCachedRaw(c.GuildID, c.DiscordID)
}

func KeyCopaingCachedRaw(guildID, copaingID string) state.Key {
	return KeyCopaingCachedPrefix + state.Key(guildID+":"+copaingID)
}

type State struct {
	mu      sync.RWMutex
	storage state.MapStorage[CopaingCached]
}

func (s *State) Copaing(guildID, copaingID string) (*CopaingCached, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	c, err := s.storage.Get(KeyCopaingCachedRaw(guildID, copaingID))
	if err != nil {
		return nil, err
	}
	mC := c.(CopaingCached)
	return &mC, nil
}

// CopaingAdd does not call Copaing.Load!
func (s *State) CopaingAdd(c *Copaing, xpToAdd uint) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sum := calcXP(c)
	var err error
	var cc *CopaingCached
	if cc, err = s.Copaing(c.GuildID, c.DiscordID); err != nil {
		cc.XPs = sum
		cc.XPToAdd = xpToAdd
	} else {
		cc = &CopaingCached{
			ID:        c.ID,
			DiscordID: c.DiscordID,
			GuildID:   c.GuildID,
			XPs:       sum,
			XPToAdd:   xpToAdd,
		}
	}
	return s.storage.Write(KeyCopaingCached(c), *cc)
}

func (s *State) CopaingRemove(c *Copaing) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.storage.Delete(KeyCopaingCached(c))
}

func calcXP(c *Copaing) uint {
	var sum uint
	for _, entry := range c.CopaingXPs {
		sum += entry.XP
	}
	return sum
}
