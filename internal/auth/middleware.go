package auth

import (
	"errors"
	"net/http"

	"github.com/dangerousmonk/short-url/internal/logging"
	"github.com/google/uuid"
)

const (
	secretKey        = "b6e2490a47c14cb7a1732aed3ba3f3c5"
	UserIDHeaderName = "x-user-id"
	AuthCookieName   = "auth"
)

func AuthMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		jwtAuthenticator, err := NewJWTAuthenticator(secretKey)
		if err != nil {
			logging.Log.Infof("Failed initialize jwtAuthenticator | %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		cookie, err := r.Cookie(AuthCookieName)

		if err != nil {
			if !errors.Is(err, http.ErrNoCookie) {
				logging.Log.Warnf("AuthMiddleware cookie err | %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			userID := uuid.New().String()
			logging.Log.Infof("AuthMiddleware err=%v, generated new userID | %v", err, userID)
			token, err := jwtAuthenticator.CreateToken(userID, tokenLifeTime)
			if err != nil {
				logging.Log.Warnf("AuthMiddleware failed generate auth token | %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			logging.Log.Infof("AuthMiddleware set cookie=%v", token)
			http.SetCookie(w, &http.Cookie{
				Name:  AuthCookieName,
				Value: token,
				Path:  "/",
			})
			r.Header.Set(UserIDHeaderName, userID)
		} else {
			claims, err := jwtAuthenticator.ValidateToken(cookie.Value)
			if err != nil {
				logging.Log.Warnf("AuthMiddleware ValidateToken error %v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			userID := claims.UserID
			logging.Log.Infof("AuthMiddleware found userID in cookie | %v", userID)
			r.Header.Set(UserIDHeaderName, userID)
		}

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
