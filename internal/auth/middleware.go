package auth

import (
	"errors"
	"net/http"

	"github.com/dangerousmonk/short-url/internal/logging"
	"github.com/google/uuid"
)

const (
	UserIDHeaderName = "x-user-id"
	AuthCookieName   = "auth"
)

func AuthMiddleware(authenticator Authenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(AuthCookieName)

			if err != nil {
				if !errors.Is(err, http.ErrNoCookie) {
					logging.Log.Warnf("AuthMiddleware cookie err | %v", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				userID := uuid.New().String()
				token, err := authenticator.CreateToken(userID, tokenLifeTime)
				if err != nil {
					logging.Log.Warnf("AuthMiddleware failed generate auth token | %v", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				http.SetCookie(w, &http.Cookie{
					Name:  AuthCookieName,
					Value: token,
					Path:  "/",
				})
				r.Header.Set(UserIDHeaderName, userID)
			} else {
				claims, err := authenticator.ValidateToken(cookie.Value)
				if err != nil {
					logging.Log.Warnf("AuthMiddleware ValidateToken error %v", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				userID := claims.UserID
				r.Header.Set(UserIDHeaderName, userID)
			}

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
