package auth

import (
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/dangerousmonk/short-url/internal/logging"
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
				err = setCookie(authenticator, w, r)
				if err != nil {
					logging.Log.Warnf("AuthMiddleware failed setCookie | %v", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			} else {
				err = resolveCookie(authenticator, r, cookie)
				if err != nil {
					logging.Log.Warnf("AuthMiddleware resolveCookie error %v", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

func setCookie(auth Authenticator, w http.ResponseWriter, r *http.Request) error {
	userID := uuid.New().String()
	token, err := auth.CreateToken(userID, tokenLifeTime)
	if err != nil {
		logging.Log.Warnf("setCookie error %v", err)
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:  AuthCookieName,
		Value: token,
		Path:  "/",
	})
	r.Header.Set(UserIDHeaderName, userID)
	return nil
}

func resolveCookie(auth Authenticator, r *http.Request, cookie *http.Cookie) error {
	claims, err := auth.ValidateToken(cookie.Value)
	if err != nil {
		return err
	}
	userID := claims.UserID
	r.Header.Set(UserIDHeaderName, userID)
	return nil
}
