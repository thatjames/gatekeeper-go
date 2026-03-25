package v1

import (
	"context"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type OIDCConfig struct {
	Provider     *oidc.Provider
	Verifier     *oidc.IDTokenVerifier
	Oauth2Config oauth2.Config
}

type OIDCOptions struct {
	IssuerURL    string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

func NewOIDCConfig(ctx context.Context, options OIDCOptions) (*OIDCConfig, error) {
	provider, err := oidc.NewProvider(ctx, options.IssuerURL)
	if err != nil {
		return nil, err
	}

	scopes := append([]string{oidc.ScopeOpenID}, options.Scopes...)
	oauth2Config := oauth2.Config{
		ClientID:     options.ClientID,
		ClientSecret: options.ClientSecret,
		RedirectURL:  options.RedirectURL,
		Scopes:       scopes,
		Endpoint:     provider.Endpoint(),
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: options.ClientID})

	return &OIDCConfig{
		Provider:     provider,
		Verifier:     verifier,
		Oauth2Config: oauth2Config,
	}, nil
}
