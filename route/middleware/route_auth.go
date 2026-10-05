package middleware

import (
	"crypto"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"gitlab.com/posfin-unigo/middleware/agen-pos/backend/gateway-service/util"
)

// Names of the route-level middlewares an admin may attach to a route.
const (
	NameTimeout        = "timeout"
	NameRetry          = "retry"
	NameCircuitBreaker = "circuit-breaker"
	NameJWT            = "jwt"
	NameAPIKey         = "api-key"
)

// KnownNames lists every middleware name accepted in Route.Middleware.
var KnownNames = []string{NameTimeout, NameRetry, NameCircuitBreaker, NameJWT, NameAPIKey}

// JWTConfig configures the "jwt" route middleware. HS256 is enabled by Secret
// and RS256 by PublicKeyPEM. With neither set the middleware rejects every request.
type JWTConfig struct {
	Secret       string
	PublicKeyPEM string
	Issuer       string
	Audience     string
}

// JWTAuth validates a Bearer JWT (HS256/RS256, exp required) and stores the
// claims in the context under util.ContextJwtClaimKey.
func JWTAuth(cfg JWTConfig) echo.MiddlewareFunc {
	var rsaKey *rsa.PublicKey
	if cfg.PublicKeyPEM != "" {
		rsaKey, _ = parseRSAPublicKey(cfg.PublicKeyPEM)
	}
	secret := []byte(cfg.Secret)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if len(secret) == 0 && rsaKey == nil {
				return echo.NewHTTPError(http.StatusServiceUnavailable, "JWT auth is not configured")
			}
			authz := c.Request().Header.Get(echo.HeaderAuthorization)
			token := strings.TrimSpace(strings.TrimPrefix(authz, "Bearer "))
			if token == "" || token == authz {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing bearer token")
			}
			claims, err := verifyJWT(token, secret, rsaKey, cfg, time.Now())
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
			}
			c.Set(util.ContextJwtClaimKey, claims)
			return next(c)
		}
	}
}

func parseRSAPublicKey(pemStr string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(strings.ReplaceAll(pemStr, `\n`, "\n")))
	if block == nil {
		return nil, errors.New("invalid PEM")
	}
	if pub, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
		if k, ok := pub.(*rsa.PublicKey); ok {
			return k, nil
		}
	}
	if cert, err := x509.ParseCertificate(block.Bytes); err == nil {
		if k, ok := cert.PublicKey.(*rsa.PublicKey); ok {
			return k, nil
		}
	}
	return nil, errors.New("not an RSA public key")
}

func verifyJWT(token string, secret []byte, rsaKey *rsa.PublicKey, cfg JWTConfig, now time.Time) (map[string]interface{}, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("malformed token")
	}
	var header struct {
		Alg string `json:"alg"`
	}
	if err := decodeSegment(parts[0], &header); err != nil {
		return nil, err
	}
	signingInput := parts[0] + "." + parts[1]
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, err
	}

	// The algorithm is chosen by what is configured, never trusted blindly
	// from the token (rejects "none" and HS/RS confusion).
	switch header.Alg {
	case "HS256":
		if len(secret) == 0 {
			return nil, errors.New("HS256 not enabled")
		}
		mac := hmac.New(sha256.New, secret)
		mac.Write([]byte(signingInput))
		if subtle.ConstantTimeCompare(mac.Sum(nil), sig) != 1 {
			return nil, errors.New("bad signature")
		}
	case "RS256":
		if rsaKey == nil {
			return nil, errors.New("RS256 not enabled")
		}
		sum := sha256.Sum256([]byte(signingInput))
		if err := rsa.VerifyPKCS1v15(rsaKey, crypto.SHA256, sum[:], sig); err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("unsupported alg")
	}

	var claims map[string]interface{}
	if err := decodeSegment(parts[1], &claims); err != nil {
		return nil, err
	}
	exp, ok := claims["exp"].(float64)
	if !ok || !now.Before(time.Unix(int64(exp), 0)) {
		return nil, errors.New("expired or missing exp")
	}
	if nbf, ok := claims["nbf"].(float64); ok && now.Before(time.Unix(int64(nbf), 0)) {
		return nil, errors.New("not yet valid")
	}
	if cfg.Issuer != "" && claims["iss"] != cfg.Issuer {
		return nil, errors.New("bad issuer")
	}
	if cfg.Audience != "" && !hasAudience(claims["aud"], cfg.Audience) {
		return nil, errors.New("bad audience")
	}
	return claims, nil
}

func decodeSegment(seg string, v interface{}) error {
	raw, err := base64.RawURLEncoding.DecodeString(seg)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, v)
}

func hasAudience(aud interface{}, want string) bool {
	switch v := aud.(type) {
	case string:
		return v == want
	case []interface{}:
		for _, a := range v {
			if s, ok := a.(string); ok && s == want {
				return true
			}
		}
	}
	return false
}

// APIKeyAuth accepts requests whose x-api-key header matches one of keys.
// With no keys configured it rejects every request.
func APIKeyAuth(keys []string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if len(keys) == 0 {
				return echo.NewHTTPError(http.StatusServiceUnavailable, "API key auth is not configured")
			}
			got := []byte(c.Request().Header.Get("x-api-key"))
			ok := false
			for _, k := range keys { // no early exit: constant work per configured key
				if subtle.ConstantTimeCompare(got, []byte(k)) == 1 {
					ok = true
				}
			}
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid API key")
			}
			return next(c)
		}
	}
}
