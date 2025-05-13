package config

import (
	"github.com/anhgelus/gokord"
	"github.com/anhgelus/gokord/utils"
	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client

func GetRedisClient() (*redis.Client, error) {
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
		utils.SendAlert("config/redis.go - Closing redis client", err.Error())
	}
}
