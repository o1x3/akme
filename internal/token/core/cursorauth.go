package core

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"strings"
)

// cursorSession holds the WorkOS session cookie pieces used by the Cursor
// dashboard usage API (same auth the web UI uses — not a crsr_ API key).
type cursorSession struct {
	Sub string // cookie id (bare user_01… / auth0|… after connection prefix strip)
	JWT string // raw access token from Cursor's local store
}

// cookieValue is WorkosCursorSessionToken=<sub>%3A%3A<jwt>.
func (s cursorSession) cookieValue() string {
	return s.Sub + "%3A%3A" + s.JWT
}

// resolveCursorSession finds a Cursor dashboard session.
// Priority: AKME_CURSOR_SESSION_TOKEN / NX_CURSOR_SESSION_TOKEN / CURSOR_SESSION_TOKEN env → state.vscdb
// ItemTable cursorAuth/accessToken.
func resolveCursorSession() (cursorSession, bool) {
	for _, key := range []string{"AKME_CURSOR_SESSION_TOKEN", "NX_CURSOR_SESSION_TOKEN", "CURSOR_SESSION_TOKEN"} {
		if v := strings.TrimSpace(os.Getenv(key)); v != "" {
			if s, ok := parseCursorSessionOverride(v); ok {
				return s, true
			}
		}
	}
	return readCursorAccessToken()
}

// parseCursorSessionOverride accepts a raw JWT or a pre-formed "sub::jwt"
// (plain or with %3A%3A).
func parseCursorSessionOverride(v string) (cursorSession, bool) {
	v = strings.TrimSpace(v)
	if v == "" {
		return cursorSession{}, false
	}
	if strings.Contains(v, "%3A%3A") {
		parts := strings.SplitN(v, "%3A%3A", 2)
		if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
			return cursorSession{Sub: cookieSub(parts[0]), JWT: parts[1]}, true
		}
	}
	if i := strings.Index(v, "::"); i > 0 {
		sub, jwt := v[:i], v[i+2:]
		if sub != "" && jwt != "" {
			return cursorSession{Sub: cookieSub(sub), JWT: jwt}, true
		}
	}
	sub, ok := jwtCookieSub(v)
	if !ok {
		return cursorSession{}, false
	}
	return cursorSession{Sub: sub, JWT: v}, true
}

func readCursorAccessToken() (cursorSession, bool) {
	for _, p := range cursorIDEPaths() {
		if _, err := os.Stat(p); err != nil {
			continue
		}
		if s, ok := readAccessTokenFromDB(p); ok {
			return s, true
		}
	}
	return cursorSession{}, false
}

func readAccessTokenFromDB(path string) (cursorSession, bool) {
	db, cleanup, err := openDB(path)
	if err != nil {
		return cursorSession{}, false
	}
	defer cleanup()

	var value string
	err = db.QueryRow(`SELECT value FROM ItemTable WHERE key = ?`, "cursorAuth/accessToken").Scan(&value)
	if err != nil || value == "" {
		return cursorSession{}, false
	}
	value = strings.TrimSpace(strings.Trim(value, `"`))
	claims, ok := jwtClaims(value)
	if !ok {
		return cursorSession{}, false
	}
	// Agent API keys authenticate api.cursor.com, not the usage dashboard.
	if strings.EqualFold(claims.Type, "api_key_token") {
		return cursorSession{}, false
	}
	sub := cookieSub(claims.Sub)
	if sub == "" {
		return cursorSession{}, false
	}
	return cursorSession{Sub: sub, JWT: value}, true
}

type cursorJWTClaims struct {
	Sub  string `json:"sub"`
	Type string `json:"type"`
}

// jwtClaims decodes the payload of a JWT.
func jwtClaims(token string) (cursorJWTClaims, bool) {
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return cursorJWTClaims{}, false
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		padded := parts[1]
		if m := len(padded) % 4; m != 0 {
			padded += strings.Repeat("=", 4-m)
		}
		payload, err = base64.URLEncoding.DecodeString(padded)
		if err != nil {
			return cursorJWTClaims{}, false
		}
	}
	var claims cursorJWTClaims
	if json.Unmarshal(payload, &claims) != nil || claims.Sub == "" {
		return cursorJWTClaims{}, false
	}
	return claims, true
}

// jwtCookieSub returns the dashboard cookie id derived from a JWT.
func jwtCookieSub(token string) (string, bool) {
	claims, ok := jwtClaims(token)
	if !ok {
		return "", false
	}
	sub := cookieSub(claims.Sub)
	return sub, sub != ""
}

// jwtSub is kept for older call sites/tests; prefer jwtCookieSub.
func jwtSub(token string) (string, bool) { return jwtCookieSub(token) }

// cookieSub strips a WorkOS connection prefix (e.g. "github|user_01…" → "user_01…").
// The dashboard cookie wants the bare id that /api/auth/me reports as sub.
func cookieSub(sub string) string {
	sub = strings.TrimSpace(sub)
	if i := strings.LastIndex(sub, "|"); i >= 0 && i+1 < len(sub) {
		return sub[i+1:]
	}
	return sub
}
