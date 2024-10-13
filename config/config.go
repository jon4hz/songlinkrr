package config

import (
	"errors"
	"os"
	"path"

	"github.com/adrg/xdg"
	"github.com/spf13/viper"
)

type PlexConfig struct {
	Username  string `mapstructure:"username"`
	Token     string `mapstructure:"token"`
	URL       string `mapstructure:"url"`
	IgnoreTLS bool   `mapstructure:"ignore_tls"`
	Timeout   int    `mapstructure:"timeout"`
}

type SubsonicConfig struct {
	URL      string `mapstructure:"url"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}

type Config struct {
	PlexConfig     PlexConfig     `mapstructure:"plex"`
	SubsonicConfig SubsonicConfig `mapstructure:"subsonic"`
}

func init() {
	viper.SetDefault("plex_timeout", 10)
}

func Load(path string) (cfg *Config, err error) {
	if path != "" {
		return load(path)
	}
	for _, f := range [...]string{
		".config.yml",
		"config.yml",
		".config.yaml",
		"config.yaml",
		"songlinkrr.yml",
		"songlinkrr.yaml",
	} {
		cfg, err = load(f)
		if err != nil && os.IsNotExist(err) {
			err = nil
			continue
		} else if err != nil && errors.As(err, &viper.ConfigFileNotFoundError{}) {
			err = nil
			continue
		}
	}
	if cfg == nil {
		return cfg, viper.Unmarshal(&cfg)
	}
	return
}

func load(file string) (cfg *Config, err error) {
	viper.SetConfigName(file)
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./")
	viper.AddConfigPath(path.Join(xdg.ConfigHome, "songlinkrr"))
	viper.AddConfigPath("/etc/songlinkrr/")
	if err = viper.ReadInConfig(); err != nil {
		return
	}
	if err = viper.Unmarshal(&cfg); err != nil {
		return
	}
	return
}
