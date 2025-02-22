package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/aymanbagabas/go-osc52/v2"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/jon4hz/songlinkrr/config"
	"github.com/jon4hz/songlinkrr/player"
	"github.com/jon4hz/songlinkrr/player/plex"
	"github.com/jon4hz/songlinkrr/version"
	"github.com/muesli/termenv"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/supersonic-app/go-subsonic/subsonic"
)

var rootCmd = &cobra.Command{
	Use:   "songlinkrr",
	Short: "Songlinkrr is a CLI tool to get song links for your currently playing song on Plex",
	Run:   root,
}

var rootCmdFlags struct {
	configFile   string
	forceConfirm bool
}

func init() {
	rootCmd.Flags().StringVarP(&rootCmdFlags.configFile, "config", "c", "", "path to the config file")
	rootCmd.Flags().BoolVarP(&rootCmdFlags.forceConfirm, "force-confirm", "f", false, "force confirmation of the song")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func root(cmd *cobra.Command, _ []string) {
	lipgloss.SetColorProfile(termenv.TrueColor)
	cfg, err := config.Load(rootCmdFlags.configFile)
	if err != nil {
		log.Fatal("Failed to load config", "err", err)
	}

	if cfg.PlexConfig.Username == "" {
		log.Fatal("No client configured")
	}

	var client player.Player
	if cfg.PlexConfig.Username != "" {
		client = plex.New(&cfg.PlexConfig, cfg.PlexConfig.Username)
	}

	var sessions []*player.Session
	if err := spinner.New().
		Type(spinner.Dots).
		Title("Fetching Plex sessions").
		Action(func() {
			var err error
			sessions, err = client.GetSessions(cmd.Context())
			if err != nil && errors.Is(err, player.ErrNoSessions) {
				log.Info(err)
				return
			} else if err != nil {
				log.Fatal("Failed to get Plex sessions", "err", err)
			}
		}).Run(); err != nil {
		log.Fatal("Failed to run spinner", "err", err)
	}

	var session *player.Session
	if rootCmdFlags.forceConfirm || len(sessions) > 1 {
		if err := huh.NewSelect[*player.Session]().
			Title("Select which song to share from Plex").
			Options(lo.Map(sessions, func(s *player.Session, _ int) huh.Option[*player.Session] {
				return huh.NewOption(
					fmt.Sprintf("%s • %s (%s)", s.Artist, s.Title, s.Player),
					s,
				)
			})...).
			Value(&session).
			Run(); err != nil {
			log.Fatal("Failed to run form", "err", err)
		}
	} else if len(sessions) == 1 {
		session = sessions[0]
	} else {
		log.Fatalf("No active music session found on %s", client.String())
	}

	subsonicClient := &subsonic.Client{
		Client:     http.DefaultClient,
		BaseUrl:    cfg.SubsonicConfig.URL,
		User:       cfg.SubsonicConfig.User,
		ClientName: "songlinkrr-" + version.Version,
	}

	if err := subsonicClient.Authenticate(cfg.SubsonicConfig.Password); err != nil {
		log.Fatal("Failed to authenticate to subsonic", "url", cfg.SubsonicConfig.URL, "err", err)
	}

	searchString := fmt.Sprintf("%s %s", session.Artist, session.Title)

RetrySearch:
	var searchResult *subsonic.SearchResult3
	if err := spinner.New().
		Type(spinner.Dots).
		Title("Searching for song on subsonic").
		Action(func() {
			var err error
			searchResult, err = subsonicClient.Search3(searchString, nil)
			if err != nil {
				log.Fatal("Failed to search for song", "err", err)
			}
		}).Run(); err != nil {
		log.Fatal("Failed to run spinner", "err", err)
	}

	if len(searchResult.Song) == 0 {
		log.Warn("No matching songs found on subsonic", "query", searchString)
		if err := huh.NewInput().
			Title("Adjust search query").
			Description("Or press ctrl+c to exit").
			Value(&searchString).
			Run(); err != nil {
			log.Fatal("Failed to get search query", "err", err)
		}
		goto RetrySearch
	}

	var song *subsonic.Child
	if err := huh.NewSelect[*subsonic.Child]().
		Title("Select best match from subsonic").
		Description(fmt.Sprintf("Search: %s", searchString)).
		Options(lo.Map(searchResult.Song, func(s *subsonic.Child, _ int) huh.Option[*subsonic.Child] {
			return huh.NewOption(fmt.Sprintf("%s • %s", s.Artist, s.Title), s)
		})...).
		Value(&song).
		Run(); err != nil {
		log.Fatal("Failed to select song", "err", err)
	}

	share, err := subsonicClient.CreateShare(song.ID, nil)
	if err != nil {
		log.Fatal("Failed to create share link", "err", err)
	}

	doCopy := true
	if err := huh.NewConfirm().
		Title("Copy share link to clipboard?").
		Affirmative("Sure!").
		Negative("Nope.").
		Value(&doCopy).
		Run(); err != nil {
		log.Fatal("Failed to confirm", "err", err)
	}
	if doCopy {
		if _, err := osc52.New(share.Url).WriteTo(os.Stderr); err != nil {
			log.Fatal("Failed to copy share link to clipboard", "err", err)
		}
	}

	fmt.Println(share.Url)
}
