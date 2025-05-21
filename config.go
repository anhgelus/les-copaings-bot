package main

import (
	"fmt"
	"github.com/anhgelus/gokord"
	"github.com/pelletier/go-toml/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	Debug    bool            `toml:"debug"`
	Author   string          `toml:"author"`
	Database *PostgresConfig `toml:"database"`
}

type PostgresConfig struct {
	Host     string `toml:"host"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	DBName   string `toml:"db_name"`
	Port     int    `toml:"port"`
}

func (p *PostgresConfig) SetDefaultValues() {
	p.Host = "localhost"
	p.User = ""
	p.Password = ""
	p.DBName = ""
	p.Port = 5432
}

func (p *PostgresConfig) Connect() (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(p.generateDsn()), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

// generateDsn for the connection to postgres using the given config.SQLCredentials
func (p *PostgresConfig) generateDsn() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Europe/Paris",
		p.Host, p.User, p.Password, p.DBName, p.Port,
	)
}

func (c *Config) IsDebug() bool {
	return c.Debug
}

func (c *Config) GetAuthor() string {
	return c.Author
}

func (c *Config) GetRedisCredentials() *gokord.RedisCredentials {
	return nil
}

func (c *Config) SetDefaultValues() {
	c.Debug = false
	c.Author = "anhgelus"
	c.Database = &PostgresConfig{}
	c.Database.SetDefaultValues()
}

func (c *Config) GetSQLCredentials() gokord.SQLCredentials {
	return c.Database
}

func (c *Config) Marshal() ([]byte, error) {
	return toml.Marshal(c)
}

func (c *Config) Unmarshal(bytes []byte) error {
	return toml.Unmarshal(bytes, c)
}
