package common

import (
	"context"

	"gorm.io/gorm"
)

type key uint8

const (
	keyDB     key = 0
	keyDebug  key = 1
	keyAuthor key = 2
)

func SetDB(ctx context.Context, db *gorm.DB) context.Context {
	return context.WithValue(ctx, keyDB, db)
}

func GetDB(ctx context.Context) *gorm.DB {
	raw := ctx.Value(keyDB)
	if raw == nil {
		return nil
	}
	return raw.(*gorm.DB)
}

func SetDebug(ctx context.Context, b bool) context.Context {
	return context.WithValue(ctx, keyDebug, b)
}

func IsDebug(ctx context.Context) bool {
	raw := ctx.Value(keyDebug)
	if raw == nil {
		return false
	}
	return raw.(bool)
}

func SetAuthor(ctx context.Context, s string) context.Context {
	return context.WithValue(ctx, keyAuthor, s)
}

func GetAuthor(ctx context.Context) string {
	raw := ctx.Value(keyAuthor)
	if raw == nil {
		return ""
	}
	return raw.(string)
}
