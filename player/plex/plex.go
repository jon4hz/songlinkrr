package plex

import (
	"context"
	"fmt"

	"github.com/jon4hz/songlinkrr/config"
	"github.com/jon4hz/songlinkrr/player"
	"github.com/jon4hz/songlinkrr/plex"
)

var _ player.Player = (*Player)(nil)

type Player struct {
	client     *plex.Client
	wantedUser string
}

func New(cfg *config.PlexConfig, wantedUser string) *Player {
	client := plex.New(
		cfg.URL,
		cfg.Token,
		cfg.Timeout,
		cfg.IgnoreTLS,
	)
	return &Player{
		client:     client,
		wantedUser: wantedUser,
	}
}

func (p *Player) String() string {
	return "Plex"
}

func (p *Player) GetSessions(ctx context.Context) ([]*player.Session, error) {
	sessions, err := p.client.GetSessions(ctx)
	if err != nil {
		return nil, err
	}

	if sessions.MediaContainer.Size == 0 {
		return nil, player.ErrNoSessions
	}

	var playerSessions []*player.Session
	for _, session := range sessions.MediaContainer.Metadata {
		if session.Type != "track" {
			continue
		}
		if p.wantedUser != "" && session.User.Title != p.wantedUser {
			continue
		}
		playerSessions = append(playerSessions, &player.Session{
			Artist: session.GrandparentTitle,
			Title:  session.Title,
			User:   session.User.Title,
			Player: fmt.Sprintf("%s %s", session.Player.Product, session.Player.Title),
		})
	}
	return playerSessions, nil
}
