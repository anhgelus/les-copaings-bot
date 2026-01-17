package user

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/nyttikord/gokord/state"
)

var ErrSyncingUnsavedData = errors.New("trying to sync unsaved data")

type CopaingCached struct {
	ID        uint
	DiscordID string
	GuildID   string
	XPs       uint
	XPToAdd   uint
	lastSync  time.Time // time.Time of the lastSync
}

// copaing turns a CopaingCached into a Copaing.
// This operation get the copaing synced with the database.
// It doesn't:
// - save the copaing in the database, use CopaingCached.SaveInDB for that;
// - sync the copaing cached, use CopaingCached.Sync for that.
// TL;DR: don't use this method, unless you know what are you doing.
func (cc *CopaingCached) copaing() *Copaing {
	c := &Copaing{DiscordID: cc.DiscordID, GuildID: cc.GuildID}
	if err := c.load(); err != nil {
		panic(err)
	}
	return c
}

func (cc *CopaingCached) Sync(ctx context.Context) error {
	if cc.mustSave() {
		return ErrSyncingUnsavedData
	}
	synced := FromCopaing(cc.copaing())
	synced.XPs += cc.XPToAdd
	synced.XPToAdd = cc.XPToAdd
	synced.lastSync = time.Now()
	err := synced.Save(ctx)
	if err != nil {
		return err
	}
	*cc = *synced
	return nil
}

func (cc *CopaingCached) Save(ctx context.Context) error {
	state := GetState(ctx)

	state.mu.Lock()
	defer state.mu.Unlock()

	return state.storage.Write(KeyCopaingCachedRaw(cc.GuildID, cc.DiscordID), *cc)
}

func (cc *CopaingCached) SaveInDB(ctx context.Context) error {
	c := cc.copaing()
	c.CopaingXPs = append(c.CopaingXPs, CopaingXP{CopaingID: c.ID, XP: cc.XPToAdd, GuildID: c.GuildID})
	err := c.Save()
	if err != nil {
		return err
	}
	cc.XPToAdd = 0
	return cc.Save(ctx)
}

func (cc *CopaingCached) Delete(ctx context.Context) error {
	c := cc.copaing()
	err := c.Delete()
	if err != nil {
		return err
	}
	state := GetState(ctx)

	state.mu.Lock()
	defer state.mu.Unlock()

	return state.storage.Delete(KeyCopaingCached(c))
}

func (cc *CopaingCached) mustSave() bool {
	return cc.XPToAdd > 0
}

func saveStateInDB(ctx context.Context) error {
	for _, v := range GetState(ctx).storage {
		if v.mustSave() {
			err := v.SaveInDB(ctx)
			if err != nil {
				return err
			}
		}
	}
	return nil
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

	raw, err := s.storage.Get(KeyCopaingCachedRaw(guildID, copaingID))
	if err != nil {
		return nil, err
	}
	c := raw.(CopaingCached)
	return &c, nil
}

func (s *State) Copaings(guild string) []CopaingCached {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var ccs []CopaingCached
	for _, cc := range s.storage {
		if cc.GuildID == guild {
			ccs = append(ccs, cc)
		}
	}
	return ccs
}

func calcXP(c *Copaing) uint {
	var sum uint
	for _, entry := range c.CopaingXPs {
		sum += entry.XP
	}
	return sum
}
