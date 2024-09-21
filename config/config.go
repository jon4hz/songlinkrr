package config

import (
	"errors"
	"os"
	"path"

	"github.com/adrg/xdg"
	"github.com/spf13/viper"
)

type Config struct {
	PlexUsername     string `mapstructure:"plex_username"`
	PlexToken        string `mapstructure:"plex_token"`
	PlexURL          string `mapstructure:"plex_url"`
	PlexIgnoreTLS    bool   `mapstructure:"plex_ignore_tls"`
	PlexTimeout      int    `mapstructure:"plex_timeout"`
	SubsonicURL      string `mapstructure:"subsonic_url"`
	SubsonicUser     string `mapstructure:"subsonic_user"`
	SubsonicPassword string `mapstructure:"subsonic_password"`
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
