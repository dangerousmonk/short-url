package auth

import (
	"net/http"

	"github.com/dangerousmonk/short-url/internal/logging"
	"github.com/google/uuid"
)

const (
	secretKey        = "b6e2490a47c14cb7a1732aed3ba3f3c5"
	UserIDHeaderName = "x-user-id"
	AuthCookieName   = "auth"
	AuthHeaderName   = "Authorization"
)

func AuthMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		jwtAuthenticator, err := NewJWTAuthenticator(secretKey)
		if err != nil {
			logging.Log.Infof("Failed initialize jwtAuthenticator | %v", err)
			http.Error(w, `{"error":" Failed initialize jwtAuthenticator"}`, http.StatusInternalServerError)
			return
		}

		var token string
		// authHeader := r.Header.Get(AuthHeaderName)
		cookie, err := r.Cookie(AuthCookieName)
		if err != nil || cookie.Value == "" {
			token = ""
		} else {
			token = cookie.Value
		}

		// switch authHeader {
		// case "":
		// 	cookie, err := r.Cookie(AuthCookieName)
		// 	if err != nil || cookie.Value == "" {
		// 		token = ""
		// 	} else {
		// 		token = cookie.Value
		// 	}
		// default:
		// 	token = authHeader
		// }

		claims, err := jwtAuthenticator.ValidateToken(token)

		var userID string

		if err != nil {
			userID = uuid.New().String()
			token, err := jwtAuthenticator.CreateToken(userID, tokenLifeTime)
			if err != nil {
				logging.Log.Infof("Failed generate auth token | %v", err)
				http.Error(w, `{"error":"Failed to generate auth token"}`, http.StatusInternalServerError)
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:     AuthCookieName,
				Value:    token,
				HttpOnly: true,
				Secure:   true,
				Path:     "/",
			})
			// w.Header().Set(AuthHeaderName, token)
		} else {
			userID = claims.UserID
			// w.Header().Set(AuthHeaderName, token)
		}

		w.Header().Set(UserIDHeaderName, userID)
		r.Header.Set(UserIDHeaderName, userID)

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
