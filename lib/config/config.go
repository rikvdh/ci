package config

import (
	"fmt"
	"github.com/go-ini/ini"
)

type Config struct {
	Dbtype string
	DbConnString string
}

var cfg Config

func getStringDefault(c *ini.File, sect string, key string, def string) string {
	if c == nil {
		return def
	}

	s := c.Section(sect).Key(key).String()
	if len(s) == 0 {
		return def
	}

	return s
}

func Load(fname string) {
	c, err := ini.LoadSources(ini.LoadOptions{AllowBooleanKeys: true}, fname)
	if err != nil {
		fmt.Println("Warning loading config:", err)
	}

	cfg.Dbtype = getStringDefault(c, "database", "type", "sqlite3")
	cfg.DbConnString = getStringDefault(c, "database", "connection", "ci.db")
}

func Get() *Config {
	return &cfg
}
