package password

import (
	"context"
	"time"

	hibp "github.com/wneessen/go-hibp"
)

type BreachChecker struct {
	client hibp.Client
}

func NewBreachChecker() *BreachChecker {
	return &BreachChecker{
		client: hibp.New(
			hibp.WithHTTPTimeout(5*time.Second),
			hibp.WithUserAgent("Calendly_GetCourse/1.0"),
		),
	}
}

func (c *BreachChecker) IsPwned(ctx context.Context, pw string) bool {
	if ctx.Err() != nil {
		return false
	}
	match, _, err := c.client.PwnedPassAPI.CheckPassword(pw)
	if err != nil {
		return false
	}
	return match.Present()
}

// In the production environment needed to be installed offline version of the HIBP database and use it for checking passwords.
