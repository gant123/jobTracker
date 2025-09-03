package services

import (
	"context"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleOAuth struct {
	Config *oauth2.Config
}

func NewGoogleOAuth() *GoogleOAuth {
	scopes := strings.Split(os.Getenv("GOOGLE_SCOPES"), ",")
	return &GoogleOAuth{
		Config: &oauth2.Config{
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
			Scopes:       scopes,
			Endpoint:     google.Endpoint,
		},
	}
}

func (g *GoogleOAuth) AuthCodeURL(state string) string {
	// request offline access (refresh token)
	return g.Config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
}

func (g *GoogleOAuth) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	return g.Config.Exchange(ctx, code)
}

// returns an *http.Client authorized with the given token
func (g *GoogleOAuth) Client(ctx context.Context, token *oauth2.Token) *http.Client {
	return g.Config.Client(ctx, token)
}
