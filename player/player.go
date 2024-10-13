package player

import (
	"context"
	"fmt"
)

var ErrNoSessions = fmt.Errorf("no active sessions")

type Session struct {
	Artist string
	Title  string
	User   string
	Player string
}

type Player interface {
	fmt.Stringer
	GetSessions(ctx context.Context) ([]*Session, error)
}
