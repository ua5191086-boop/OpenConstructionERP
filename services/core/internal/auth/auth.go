package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
)

// ============================================================
// Keycloak JWT Auth Middleware & RBAC for Go API
// ============================================================

// CtxKey type for context keys
type CtxKey string

const (
	CtxUserID   CtxKey = "user_id"
	CtxUsername CtxKey = "username"
	CtxRoles    CtxKey = "roles"
	CtxEmail    CtxKey = "email"
	CtxToken    CtxKey = "token"
)

// KeycloakConfig holds the realm configuration
type KeycloakConfig struct {
	Realm   string
	Issuer  string
	JWKSURL string
	// For simplicity, we also support verification via JWKS fetch
	// In production, cache the JWKS and rotate periodically
}

// NewKeycloakConfig creates a default config
func NewKeycloakConfig() *KeycloakConfig {
	return &KeycloakConfig{
		Realm:   "OpenConstructionERP",
		Issuer:  "http://localhost:8084/realms/OpenConstructionERP",
		JWKSURL: "http://localhost:8084/realms/OpenConstructionERP/protocol/openid-connect/certs",
	}
}

// Claims represents the Keycloak JWT claims
type Claims struct {
	jwt.RegisteredClaims
	RealmAccess struct {
		Roles []string `json:"roles"`
	} `json:"realm_access"`
	ResourceAccess map[string]struct {
		Roles []string `json:"roles"`
	} `json:"resource_access"`
	PreferredUsername string `json:"preferred_username"`
	Email             string `json:"email"`
}

// UserInfo extracted from JWT
type UserInfo struct {
	UserID   string
	Username string
	Email    string
	Roles    []string
}

// GetUserInfo extracts user info from request context
func GetUserInfo(ctx context.Context) *UserInfo {
	return &UserInfo{
		UserID:   ctx.Value(CtxUserID).(string),
		Username: ctx.Value(CtxUsername).(string),
		Email:    ctx.Value(CtxEmail).(string),
		Roles:    ctx.Value(CtxRoles).([]string),
	}
}

// --- In-memory JWKS cache with lazy loading ---

type jwkKey struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
	Alg string `json:"alg"`
}

type jwksResponse struct {
	Keys []jwkKey `json:"keys"`
}

// --- Middleware ---

// JWTAuthMiddleware validates Keycloak JWT tokens
func JWTAuthMiddleware(cfg *KeycloakConfig) func(http.Handler) http.Handler {
	// Cache the public key after first fetch
	var cachedKey *rsa.PublicKey
	var cachedKid string
	var lastFetch time.Time

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				http.Error(w, `{"error":"invalid authorization header format"}`, http.StatusUnauthorized)
				return
			}
			tokenString := parts[1]

			// Parse token (without verification first to get kid)
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}

				kid, ok := token.Header["kid"].(string)
				if !ok {
					return nil, fmt.Errorf("missing kid in token header")
				}

				// Cache check: fetch fresh JWKS if stale or kid changed
				if time.Since(lastFetch) > 5*time.Minute || cachedKid != kid || cachedKey == nil {
					key, err := fetchPublicKey(cfg.JWKSURL, kid)
					if err != nil {
						return nil, fmt.Errorf("failed to fetch public key: %w", err)
					}
					cachedKey = key
					cachedKid = kid
					lastFetch = time.Now()
				}

				return cachedKey, nil
			})
			if err != nil || !token.Valid {
				log.Printf("[Auth] JWT validation failed: %v", err)
				http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, `{"error":"invalid token claims"}`, http.StatusUnauthorized)
				return
			}

			// Extract user info from claims
			userID, _ := claims["sub"].(string)
			username, _ := claims["preferred_username"].(string)
			email, _ := claims["email"].(string)

			var roles []string
			if ra, ok := claims["realm_access"].(map[string]interface{}); ok {
				if rArr, ok := ra["roles"].([]interface{}); ok {
					for _, r := range rArr {
						if s, ok := r.(string); ok {
							roles = append(roles, s)
						}
					}
				}
			}

			// Also check resource_access.oce-api.roles
			if ra, ok := claims["resource_access"].(map[string]interface{}); ok {
				if oceAPI, ok := ra["oce-api"].(map[string]interface{}); ok {
					if rArr, ok := oceAPI["roles"].([]interface{}); ok {
						for _, r := range rArr {
							if s, ok := r.(string); ok {
								roles = append(roles, s)
							}
						}
					}
				}
			}

			// Set user context
			ctx := context.WithValue(r.Context(), CtxUserID, userID)
			ctx = context.WithValue(ctx, CtxUsername, username)
			ctx = context.WithValue(ctx, CtxEmail, email)
			ctx = context.WithValue(ctx, CtxRoles, roles)
			ctx = context.WithValue(ctx, CtxToken, tokenString)

			// Set PostgreSQL session variables for audit logging
			if username != "" {
				setPGUserContext(ctx, r, userID, username)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequiredRoleMiddleware checks that the user has at least one of the required roles
func RequiredRoleMiddleware(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRoles, ok := r.Context().Value(CtxRoles).([]string)
			if !ok || len(userRoles) == 0 {
				http.Error(w, `{"error":"forbidden: no roles assigned"}`, http.StatusForbidden)
				return
			}

			roleSet := make(map[string]bool)
			for _, r := range userRoles {
				roleSet[r] = true
			}

			for _, required := range roles {
				if roleSet[required] {
					next.ServeHTTP(w, r)
					return
				}
			}

			http.Error(w, fmt.Sprintf(`{"error":"forbidden: requires one of roles %v"}`, roles), http.StatusForbidden)
		})
	}
}

// --- helpers ---

func fetchPublicKey(jwksURL, kid string) (*rsa.PublicKey, error) {
	resp, err := http.Get(jwksURL)
	if err != nil {
		return nil, fmt.Errorf("jwks fetch error: %w", err)
	}
	defer resp.Body.Close()

	var jwks jwksResponse
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("jwks decode error: %w", err)
	}

	for _, key := range jwks.Keys {
		if key.Kid == kid {
			return rsaPublicKeyFromJWK(key.N, key.E)
		}
	}

	return nil, fmt.Errorf("key with kid %s not found in JWKS", kid)
}

func rsaPublicKeyFromJWK(nBase64, eBase64 string) (*rsa.PublicKey, error) {
	// Decode modulus N
	nBytes, err := base64.RawURLEncoding.DecodeString(nBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode modulus: %w", err)
	}
	n := new(big.Int).SetBytes(nBytes)

	// Decode exponent E
	eBytes, err := base64.RawURLEncoding.DecodeString(eBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exponent: %w", err)
	}

	// exponent is typically 3 or 4 bytes, little-endian in JWK
	var e int
	if len(eBytes) < 8 {
		// pad to 8 bytes for binary.BigEndian
		padded := make([]byte, 8-len(eBytes))
		padded = append(padded, eBytes...)
		e = int(binary.BigEndian.Uint64(padded))
	} else {
		e = int(binary.BigEndian.Uint64(eBytes[:8]))
	}

	return &rsa.PublicKey{N: n, E: e}, nil
}

// setPGUserContext sets PostgreSQL session variables for audit logging
func setPGUserContext(ctx context.Context, r *http.Request, userID, username string) {
	// This is a lightweight approach — in production, use a db hook
	// that calls SET LOCAL on each transaction.
	// For now, we log it and let each handler set the pg context explicitly.
	log.Printf("[Auth] User context: id=%s username=%s", userID, username)
}

// RequireAuth is a convenience wrapper for chi groups
func RequireAuth(cfg *KeycloakConfig) chi.Middlewares {
	return chi.Middlewares{
		JWTAuthMiddleware(cfg),
	}
}

// AuthHandler returns user info from the request context
func AuthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := r.Context().Value(CtxUserID).(string)
		username, _ := r.Context().Value(CtxUsername).(string)
		email, _ := r.Context().Value(CtxEmail).(string)
		roles, _ := r.Context().Value(CtxRoles).([]string)
		token, _ := r.Context().Value(CtxToken).(string)

		resp := map[string]interface{}{
			"user_id":            userID,
			"username":           username,
			"email":              email,
			"roles":              roles,
			"authenticated":      userID != "",
			"token_preview":      token[:min(len(token), 20)] + "...",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

// PublicAuthMiddleware is an optional auth that extracts user info if token present,
// but does not reject unauthenticated requests
func PublicAuthMiddleware(cfg *KeycloakConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				// No token — set empty context, continue
				ctx := context.WithValue(r.Context(), CtxUserID, "")
				ctx = context.WithValue(ctx, CtxUsername, "")
				ctx = context.WithValue(ctx, CtxEmail, "")
				ctx = context.WithValue(ctx, CtxRoles, []string{})
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Delegate to full auth middleware
			JWTAuthMiddleware(cfg)(next).ServeHTTP(w, r)
		})
	}
}