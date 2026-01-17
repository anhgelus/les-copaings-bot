package user

import (
	"context"
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

// Copaing turns a CopaingCached into a Copaing.
// This operation is heavy.
func (cc *CopaingCached) Copaing(ctx context.Context) *Copaing {
	c := Copaing{DiscordID: cc.DiscordID, GuildID: cc.GuildID}
	if err := c.Load(ctx); err != nil {
		panic(err)
	}
	return &c
}

func (cc *CopaingCached) Save(ctx context.Context) error {
	state := GetState(ctx)

	state.mu.Lock()
	defer state.mu.Unlock()

	return state.storage.Write(KeyCopaingCachedRaw(cc.GuildID, cc.DiscordID), *cc)
}

func FromCopaing(c *Copaing) *CopaingCached {
	return &CopaingCached{
		ID:        c.ID,
		DiscordID: c.DiscordID,
		GuildID:   c.GuildID,
		XPs:       calcXP(c),
		XPToAdd:   0,
	}
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

func NewState() *State {
	return &State{
		storage: state.MapStorage[CopaingCached]{},
	}
}

const ContextKeyState = "state"

func GetState(ctx context.Context) *State {
	return ctx.Value(ContextKeyState).(*State)
}

func SetState(ctx context.Context, state *State) context.Context {
	return context.WithValue(ctx, ContextKeyState, state)
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
func (s *State) CopaingAdd(c *Copaing, xpToAdd uint) (*CopaingCached, error) {
	var err error
	var cc *CopaingCached
	if cc, err = s.Copaing(c.GuildID, c.DiscordID); err == nil {
		cc.XPs = calcXP(c)
		cc.XPToAdd = xpToAdd
	} else {
		cc = FromCopaing(c)
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	return cc, s.storage.Write(KeyCopaingCached(c), *cc)
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
