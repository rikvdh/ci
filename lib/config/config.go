package config

import (
	"fmt"

	"github.com/go-ini/ini"
)

// Config structure containing all configuration options
type Config struct {
	Dbtype           string
	DbConnString     string
	ConcurrentBuilds uint
	BuildDir         string
	ListeningURI     string
}

var cfg Config

func getStringDefault(c *ini.File, sect, key, def string) string {
	if c == nil {
		return def
	}

	s := c.Section(sect).Key(key).String()
	if len(s) == 0 {
		return def
	}

	return s
}

func getUintDefault(c *ini.File, sect, key string, def uint) uint {
	if c == nil {
		return def
	}

	s, err := c.Section(sect).Key(key).Uint()
	if err != nil {
		return def
	}

	return s
}

// Load loads the configuration from the ini-file
func Load(fname string) {
	c, err := ini.LoadSources(ini.LoadOptions{AllowBooleanKeys: true}, fname)
	if err != nil {
		fmt.Println("problem loading ini-file:", err)
	}

	cfg.Dbtype = getStringDefault(c, "database", "type", "sqlite3")
	cfg.DbConnString = getStringDefault(c, "database", "connection", "ci.db")
	cfg.ConcurrentBuilds = getUintDefault(c, "build", "parallel", 5)
	cfg.BuildDir = getStringDefault(c, "build", "work_dir", "./ci-build")
	cfg.ListeningURI = getStringDefault(c, "http", "listen_uri", ":8081")
}

// Get retrieves the configuration structure
func Get() *Config {
	return &cfg
}
