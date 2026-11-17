package orders

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ErrUnknownUser means users-svc answered 404 for the user ID.
var ErrUnknownUser = errors.New("unknown user")

// UserValidator checks that a user exists.
type UserValidator interface {
	ValidateUser(ctx context.Context, id string) error
}

// UsersClient calls users-svc over HTTP.
type UsersClient struct {
	baseURL string
	http    *http.Client
}

// NewUsersClient returns a client for users-svc at baseURL (e.g. http://users-svc).
func NewUsersClient(baseURL string) *UsersClient {
	return &UsersClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: 3 * time.Second},
	}
}

// ValidateUser returns nil if the user exists, ErrUnknownUser on 404, or another error
// if users-svc can't be reached.
func (c *UsersClient) ValidateUser(ctx context.Context, id string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/users/"+url.PathEscape(id), nil) //nolint:gosec // G704: base URL is operator config (USERS_SVC_URL), id is path-escaped
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req) //nolint:gosec // G704: see above
	if err != nil {
		return fmt.Errorf("call users-svc: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
		return ErrUnknownUser
	default:
		return fmt.Errorf("users-svc returned %d", resp.StatusCode)
	}
}
