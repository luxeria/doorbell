package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/luxeria/doorbell/pkg/jwt"
	"github.com/luxeria/doorbell/pkg/recaptcha"
	"github.com/luxeria/doorbell/pkg/rest"
)

type Config struct {
	JwtSecret         []byte
	JwtExpiry         time.Duration
	Recaptcha         *recaptcha.Recaptcha
	RecaptchaMinScore float64
}

type Auth struct {
	jwtSecret         []byte
	jwtExpiry         time.Duration
	recaptcha         *recaptcha.Recaptcha
	recaptchaMinScore float64
}

func New(c Config) *Auth {
	if len(c.JwtSecret) == 0 {
		panic("jwt secret must not be empty")
	}

	if c.JwtExpiry == 0 {
		panic("jwt expiration time must not be zero")
	}

	if c.Recaptcha == nil {
		panic("recaptcha object must not be nil")
	}

	if !(c.RecaptchaMinScore > 0.0 && c.RecaptchaMinScore < 1.0) {
		panic("recaptcha min score must be between 0.0 and 1.0")
	}

	return &Auth{
		jwtSecret:         c.JwtSecret,
		jwtExpiry:         c.JwtExpiry,
		recaptcha:         c.Recaptcha,
		recaptchaMinScore: c.RecaptchaMinScore,
	}
}

type authRecaptchaRequest struct {
	Response string `json:"response"`
}

type authResponse struct {
	Token string `json:"token"`
}

func (a *Auth) AuthRecaptcha() http.Handler {
	return rest.PostRequest(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// decode request
		var req authRecaptchaRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			rest.Error(w, r, err, http.StatusBadRequest)
			return
		}

		// check and verify recaptcha score
		v, err := a.recaptcha.Verify(req.Response)
		if err != nil {
			rest.Error(w, r, err, http.StatusBadRequest)
			return
		}

		if v.Score < a.recaptchaMinScore {
			rest.Error(w, r, fmt.Errorf("recaptcha score (%f) too low", v.Score), http.StatusUnauthorized)
			return
		}

		// generate jwt token
		claims := jwt.Claims{
			Subject:   "Anonymous (reCAPTCHA)",
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(a.jwtExpiry).Unix(),
		}

		token, err := jwt.Sign(claims, a.jwtSecret)
		if err != nil {
			rest.Error(w, r, err, http.StatusInternalServerError)
			return
		}

		rest.JSON(w, authResponse{Token: token}, http.StatusOK)
	}))
}

func (a *Auth) CheckJwt(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const prefix = "Bearer "

		var token string
		authorization := r.Header.Get("Authorization")
		if len(authorization) > len(prefix) && strings.EqualFold(prefix, authorization[0:len(prefix)]) {
			token = strings.TrimSpace(authorization[len(prefix):])
		}

		_, err := jwt.Verify(token, a.jwtSecret)
		if err != nil {
			rest.Error(w, r, err, http.StatusUnauthorized)
		} else {
			h.ServeHTTP(w, r)
		}
	})
}
