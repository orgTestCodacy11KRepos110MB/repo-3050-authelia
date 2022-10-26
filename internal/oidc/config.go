package oidc

import (
	"context"
	"crypto/sha512"
	"hash"
	"html/template"
	"net/url"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/oauth2"
	"github.com/ory/fosite/handler/openid"
	"github.com/ory/fosite/i18n"
	"github.com/ory/fosite/token/hmac"
	"github.com/ory/fosite/token/jwt"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

func NewConfig(config *schema.OpenIDConnectConfiguration) *Config {
	c := &Config{
		GlobalSecret:               []byte(utils.HashSHA256FromString(config.HMACSecret)),
		SendDebugMessagesToClients: config.EnableClientDebugMessages,
		MinParameterEntropy:        config.MinimumParameterEntropy,
		Lifespans: LifespanConfig{
			AccessToken:   config.AccessTokenLifespan,
			AuthorizeCode: config.AuthorizeCodeLifespan,
			IDToken:       config.IDTokenLifespan,
			RefreshToken:  config.RefreshTokenLifespan,
		},
		ProofKeyCodeExchange: ProofKeyCodeExchangeConfig{
			Enforce:                   config.EnforcePKCE == "always",
			EnforcePublicClients:      config.EnforcePKCE != "never",
			AllowPlainChallengeMethod: config.EnablePKCEPlainChallenge,
		},
	}

	c.Strategy.Core = &oauth2.HMACSHAStrategy{
		Enigma: &hmac.HMACStrategy{Config: c},
		Config: c,
	}

	return c
}

type Config struct {
	// GlobalSecret is the global secret used to sign and verify signatures.
	GlobalSecret []byte

	// RotatedGlobalSecrets is a list of global secrets that are used to verify signatures.
	RotatedGlobalSecrets [][]byte

	Issuers IssuersConfig

	SendDebugMessagesToClients    bool
	DisableRefreshTokenValidation bool
	OmitRedirectScopeParameter    bool

	JWTScopeField  jwt.JWTScopeFieldEnum
	JWTMaxDuration time.Duration

	Hash                 HashConfig
	Strategy             StrategyConfig
	PAR                  PARConfig
	Handlers             HandlersConfig
	Lifespans            LifespanConfig
	ProofKeyCodeExchange ProofKeyCodeExchangeConfig
	GrantTypeJWTBearer   GrantTypeJWTBearerConfig

	TokenURL            string
	TokenEntropy        int
	MinParameterEntropy int

	SanitationWhiteList []string
	AllowedPrompts      []string
	RefreshTokenScopes  []string

	HTTPClient           *retryablehttp.Client
	FormPostHTMLTemplate *template.Template
	MessageCatalog       i18n.MessageCatalog
}

type HashConfig struct {
	ClientSecrets fosite.Hasher
	HMAC          func() (h hash.Hash)
}

type StrategyConfig struct {
	Core                 oauth2.CoreStrategy
	OpenID               openid.OpenIDConnectTokenStrategy
	Audience             fosite.AudienceMatchingStrategy
	Scope                fosite.ScopeStrategy
	JWKSFetcher          fosite.JWKSFetcherStrategy
	ClientAuthentication fosite.ClientAuthenticationStrategy
}

type PARConfig struct {
	Enforced        bool
	URIPrefix       string
	ContextLifespan time.Duration
}

type IssuersConfig struct {
	IDToken     string
	AccessToken string
}

type HandlersConfig struct {
	// ResponseMode provides an extension handler for custom response modes.
	ResponseMode fosite.ResponseModeHandler

	// AuthorizeEndpoint is a list of handlers that are called before the authorization endpoint is served.
	AuthorizeEndpoint fosite.AuthorizeEndpointHandlers

	// TokenEndpoint is a list of handlers that are called before the token endpoint is served.
	TokenEndpoint fosite.TokenEndpointHandlers

	// TokenIntrospection is a list of handlers that are called before the token introspection endpoint is served.
	TokenIntrospection fosite.TokenIntrospectionHandlers

	// Revocation is a list of handlers that are called before the revocation endpoint is served.
	Revocation fosite.RevocationHandlers

	// PushedAuthorizeEndpoint is a list of handlers that are called before the PAR endpoint is served.
	PushedAuthorizeEndpoint fosite.PushedAuthorizeEndpointHandlers
}

type GrantTypeJWTBearerConfig struct {
	OptionalClientAuth bool
	OptionalJTIClaim   bool
	OptionalIssuedDate bool
}

type ProofKeyCodeExchangeConfig struct {
	Enforce                   bool
	EnforcePublicClients      bool
	AllowPlainChallengeMethod bool
}

type LifespanConfig struct {
	AccessToken   time.Duration
	AuthorizeCode time.Duration
	IDToken       time.Duration
	RefreshToken  time.Duration
}

const (
	PromptNone    = none
	PromptLogin   = "login"
	PromptConsent = "consent"
)

// GetAllowedPrompts returns the allowed prompts.
func (c *Config) GetAllowedPrompts(ctx context.Context) (prompts []string) {
	if len(c.AllowedPrompts) == 0 {
		return []string{PromptNone, PromptLogin, PromptConsent}
	}

	return c.AllowedPrompts
}

// GetEnforcePKCE returns the enforcement of PKCE.
func (c *Config) GetEnforcePKCE(ctx context.Context) (enforce bool) {
	return c.ProofKeyCodeExchange.Enforce
}

// GetEnforcePKCEForPublicClients returns the enforcement of PKCE for public clients.
func (c *Config) GetEnforcePKCEForPublicClients(ctx context.Context) (enforce bool) {
	return c.GetEnforcePKCE(ctx) || c.ProofKeyCodeExchange.EnforcePublicClients
}

// GetEnablePKCEPlainChallengeMethod returns the enable PKCE plain challenge method.
func (c *Config) GetEnablePKCEPlainChallengeMethod(ctx context.Context) (enable bool) {
	return c.ProofKeyCodeExchange.AllowPlainChallengeMethod
}

// GetGrantTypeJWTBearerCanSkipClientAuth returns the grant type JWT bearer can skip client auth.
func (c *Config) GetGrantTypeJWTBearerCanSkipClientAuth(ctx context.Context) (skip bool) {
	return c.GrantTypeJWTBearer.OptionalClientAuth
}

// GetGrantTypeJWTBearerIDOptional returns the grant type JWT bearer ID optional.
func (c *Config) GetGrantTypeJWTBearerIDOptional(ctx context.Context) (optional bool) {
	return c.GrantTypeJWTBearer.OptionalJTIClaim
}

// GetGrantTypeJWTBearerIssuedDateOptional returns the grant type JWT bearer issued date optional.
func (c *Config) GetGrantTypeJWTBearerIssuedDateOptional(ctx context.Context) (optional bool) {
	return c.GrantTypeJWTBearer.OptionalIssuedDate
}

// GetJWTMaxDuration returns the JWT max duration.
func (c *Config) GetJWTMaxDuration(ctx context.Context) (duration time.Duration) {
	if c.JWTMaxDuration == 0 {
		c.JWTMaxDuration = time.Hour * 24
	}

	return c.JWTMaxDuration
}

// GetRedirectSecureChecker returns the redirect URL security validator.
func (c *Config) GetRedirectSecureChecker(ctx context.Context) func(context.Context, *url.URL) (secure bool) {
	return fosite.IsRedirectURISecure
}

// GetOmitRedirectScopeParam must be set to true if the scope query param is to be omitted
// in the authorization's redirect URI.
func (c *Config) GetOmitRedirectScopeParam(ctx context.Context) (omit bool) {
	return c.OmitRedirectScopeParameter
}

// GetSanitationWhiteList is a whitelist of form values that are required by the token endpoint. These values
// are safe for storage in a database (cleartext).
func (c *Config) GetSanitationWhiteList(ctx context.Context) (whitelist []string) {
	return c.SanitationWhiteList
}

// GetJWTScopeField returns the JWT scope field.
func (c *Config) GetJWTScopeField(ctx context.Context) (field jwt.JWTScopeFieldEnum) {
	if c.JWTScopeField == jwt.JWTScopeFieldUnset {
		c.JWTScopeField = jwt.JWTScopeFieldList
	}

	return c.JWTScopeField
}

// GetIDTokenIssuer returns the ID token issuer.
func (c *Config) GetIDTokenIssuer(ctx context.Context) (issuer string) {
	return c.Issuers.IDToken
}

// GetAccessTokenIssuer returns the access token issuer.
func (c *Config) GetAccessTokenIssuer(ctx context.Context) (issuer string) {
	return c.Issuers.AccessToken
}

// GetDisableRefreshTokenValidation returns the disable refresh token validation flag.
func (c *Config) GetDisableRefreshTokenValidation(ctx context.Context) (disable bool) {
	return c.DisableRefreshTokenValidation
}

// GetAuthorizeCodeLifespan returns the authorization code lifespan.
func (c *Config) GetAuthorizeCodeLifespan(ctx context.Context) (lifespan time.Duration) {
	if c.Lifespans.AuthorizeCode <= 0 {
		return lifespanAuthorizeCodeDefault
	}

	return c.Lifespans.AuthorizeCode
}

// GetRefreshTokenLifespan returns the refresh token lifespan.
func (c *Config) GetRefreshTokenLifespan(ctx context.Context) (lifespan time.Duration) {
	if c.Lifespans.RefreshToken <= 0 {
		return lifespanRefreshTokenDefault
	}

	return c.Lifespans.RefreshToken
}

// GetIDTokenLifespan returns the ID token lifespan.
func (c *Config) GetIDTokenLifespan(ctx context.Context) (lifespan time.Duration) {
	if c.Lifespans.IDToken <= 0 {
		return lifespanTokenDefault
	}

	return c.Lifespans.IDToken
}

// GetAccessTokenLifespan returns the access token lifespan.
func (c *Config) GetAccessTokenLifespan(ctx context.Context) (lifespan time.Duration) {
	if c.Lifespans.AccessToken <= 0 {
		return lifespanTokenDefault
	}

	return c.Lifespans.AccessToken
}

// GetTokenEntropy returns the token entropy.
func (c *Config) GetTokenEntropy(ctx context.Context) (entropy int) {
	if c.TokenEntropy == 0 {
		c.TokenEntropy = 32
	}

	return c.TokenEntropy
}

// GetGlobalSecret returns the global secret.
func (c *Config) GetGlobalSecret(ctx context.Context) (secret []byte) {
	return c.GlobalSecret
}

// GetRotatedGlobalSecrets returns the rotated global secrets.
func (c *Config) GetRotatedGlobalSecrets(ctx context.Context) (secrets [][]byte) {
	return c.RotatedGlobalSecrets
}

// GetHTTPClient returns the HTTP client provider.
func (c *Config) GetHTTPClient(ctx context.Context) (client *retryablehttp.Client) {
	if c.HTTPClient == nil {
		return retryablehttp.NewClient()
	}

	return c.HTTPClient
}

// GetRefreshTokenScopes returns the refresh token scopes.
func (c *Config) GetRefreshTokenScopes(ctx context.Context) (scopes []string) {
	if c.RefreshTokenScopes == nil {
		return []string{ScopeOffline, ScopeOfflineAccess}
	}

	return c.RefreshTokenScopes
}

// GetScopeStrategy returns the scope strategy.
func (c *Config) GetScopeStrategy(ctx context.Context) (strategy fosite.ScopeStrategy) {
	if c.Strategy.Scope == nil {
		c.Strategy.Scope = fosite.ExactScopeStrategy
	}

	return c.Strategy.Scope
}

// GetAudienceStrategy returns the audience strategy.
func (c *Config) GetAudienceStrategy(ctx context.Context) (strategy fosite.AudienceMatchingStrategy) {
	if c.Strategy.Audience == nil {
		c.Strategy.Audience = fosite.DefaultAudienceMatchingStrategy
	}

	return c.Strategy.Audience
}

// GetMinParameterEntropy returns the minimum parameter entropy.
func (c *Config) GetMinParameterEntropy(_ context.Context) (entropy int) {
	if c.MinParameterEntropy == 0 {
		c.MinParameterEntropy = fosite.MinParameterEntropy
	}

	return c.MinParameterEntropy
}

// GetHMACHasher returns the hash function.
func (c *Config) GetHMACHasher(ctx context.Context) func() (h hash.Hash) {
	if c.Hash.HMAC == nil {
		c.Hash.HMAC = sha512.New512_256
	}

	return c.Hash.HMAC
}

// GetSendDebugMessagesToClients returns the send debug messages to clients.
func (c *Config) GetSendDebugMessagesToClients(ctx context.Context) (send bool) {
	return c.SendDebugMessagesToClients
}

// GetJWKSFetcherStrategy returns the JWKS fetcher strategy.
func (c *Config) GetJWKSFetcherStrategy(ctx context.Context) (strategy fosite.JWKSFetcherStrategy) {
	if c.Strategy.JWKSFetcher == nil {
		c.Strategy.JWKSFetcher = fosite.NewDefaultJWKSFetcherStrategy()
	}

	return c.Strategy.JWKSFetcher
}

// GetClientAuthenticationStrategy returns the client authentication strategy.
func (c *Config) GetClientAuthenticationStrategy(ctx context.Context) (strategy fosite.ClientAuthenticationStrategy) {
	return c.Strategy.ClientAuthentication
}

// GetMessageCatalog returns the message catalog.
func (c *Config) GetMessageCatalog(ctx context.Context) (catalog i18n.MessageCatalog) {
	return c.MessageCatalog
}

// GetFormPostHTMLTemplate returns the form post HTML template.
func (c *Config) GetFormPostHTMLTemplate(ctx context.Context) (tmpl *template.Template) {
	return c.FormPostHTMLTemplate
}

// GetTokenURL returns the token URL.
func (c *Config) GetTokenURL(ctx context.Context) (tokenURL string) {
	return c.TokenURL
}

// GetSecretsHasher returns the client secrets hashing function.
func (c *Config) GetSecretsHasher(ctx context.Context) (hasher fosite.Hasher) {
	if c.Hash.ClientSecrets == nil {
		c.Hash.ClientSecrets = &AdaptiveHasher{}
	}

	return c.Hash.ClientSecrets
}

// GetUseLegacyErrorFormat returns whether to use the legacy error format.
//
// DEPRECATED: Do not use this flag anymore.
func (c *Config) GetUseLegacyErrorFormat(ctx context.Context) (use bool) {
	return false
}

// GetAuthorizeEndpointHandlers returns the authorize endpoint handlers.
func (c *Config) GetAuthorizeEndpointHandlers(ctx context.Context) (handlers fosite.AuthorizeEndpointHandlers) {
	return c.Handlers.AuthorizeEndpoint
}

// GetTokenEndpointHandlers returns the token endpoint handlers.
func (c *Config) GetTokenEndpointHandlers(ctx context.Context) (handlers fosite.TokenEndpointHandlers) {
	return c.Handlers.TokenEndpoint
}

// GetTokenIntrospectionHandlers returns the token introspection handlers.
func (c *Config) GetTokenIntrospectionHandlers(ctx context.Context) (handlers fosite.TokenIntrospectionHandlers) {
	return c.Handlers.TokenIntrospection
}

// GetRevocationHandlers returns the revocation handlers.
func (c *Config) GetRevocationHandlers(ctx context.Context) (handlers fosite.RevocationHandlers) {
	return c.Handlers.Revocation
}

// GetPushedAuthorizeEndpointHandlers returns the handlers.
func (c *Config) GetPushedAuthorizeEndpointHandlers(ctx context.Context) fosite.PushedAuthorizeEndpointHandlers {
	return c.Handlers.PushedAuthorizeEndpoint
}

// GetResponseModeHandlerExtension returns the response mode handler extension.
func (c *Config) GetResponseModeHandlerExtension(ctx context.Context) (handler fosite.ResponseModeHandler) {
	return c.Handlers.ResponseMode
}

// GetPushedAuthorizeRequestURIPrefix is the request URI prefix. This is
// usually 'urn:ietf:params:oauth:request_uri:'.
func (c *Config) GetPushedAuthorizeRequestURIPrefix(ctx context.Context) string {
	if c.PAR.URIPrefix == "" {
		c.PAR.URIPrefix = urnPARPrefix
	}

	return c.PAR.URIPrefix
}

// EnforcePushedAuthorize indicates if PAR is enforced. In this mode, a client
// cannot pass authorize parameters at the 'authorize' endpoint. The 'authorize' endpoint
// must contain the PAR request_uri.
func (c *Config) EnforcePushedAuthorize(ctx context.Context) bool {
	return c.PAR.Enforced
}

// GetPushedAuthorizeContextLifespan is the lifespan of the short-lived PAR context.
func (c *Config) GetPushedAuthorizeContextLifespan(ctx context.Context) (lifespan time.Duration) {
	if c.PAR.ContextLifespan == 0 {
		return lifespanPARContextDefault
	}

	return c.PAR.ContextLifespan
}
