package password

import (
	"context"
	"time"

	hibp "github.com/wneessen/go-hibp"
)

type BreachChecker struct {
	client *hibp.Client
}

func NewBreachChecker() *BreachChecker {
	return &BreachChecker{
		client: hibp.New(
			hibp.WithTimeout(5 * time.Second),
			hibp.WithUserAgent("Calendly_GetCourse/1.0"),
		),
	}
}

func (c *BreachChecker) IsPwned(ctx context.Context, pw string) bool {
	pwned, err := c.client.PwnedPassAPI.CheckPassword(ctx, pw)
	if err != nil {
		return false
	}
	return pwned
}

// In the production environment needed to be installed offline version of the HIBP database and use it for checking passwords.