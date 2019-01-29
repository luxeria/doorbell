package main

import (
	"log"
	"net/http"

	"github.com/luxeria/doorbell/pkg/env"
	"github.com/luxeria/doorbell/pkg/rest/auth"
	"github.com/luxeria/doorbell/pkg/rest/doorbell"
	"github.com/luxeria/doorbell/pkg/webui"
)

func main() {
	webUi, err := webui.New("assets/webui/", webui.Context{
		"index.html": webui.Values{
			"RecaptchaSiteKey": env.String("RECAPTCHA_SITE_KEY"),
		},
	})
	if err != nil {
		log.Fatalf("failed to load webui: %s", err)
	}

	authApi := auth.New(auth.Config{
		JwtSecret:         env.Bytes("JWT_SECRET"),
		JwtExpiry:         env.Duration("JWT_EXPIRY", "15m"),
		Recaptcha:         env.Recaptcha("RECAPTCHA_SECRET_KEY"),
		RecaptchaMinScore: env.Float("RECAPTCHA_MIN_SCORE", "0.5"),
	})

	bellApi := doorbell.New(doorbell.Config{
		OpeningHours: env.OpeningHours("OPENING_HOURS", "Mo-Su 00:00-00:00"),
		RateLimit:    env.RateLimit("RATELIMIT_BURST", "3/10s"),
		DoorbellCmd:  env.StringSlice("DOORBELL_CMD", `["mpv", "assets/dingdong.mp3"]`),
	})

	http.Handle("/webui/", http.StripPrefix("/webui/", webUi))
	http.Handle("/auth/recaptcha", authApi.AuthRecaptcha())
	http.Handle("/ring", authApi.CheckJwt(bellApi.Ring()))
	http.Handle("/", http.RedirectHandler("/webui/", http.StatusFound))

	log.Fatalln(http.ListenAndServe(env.Addr("PORT", "8080"), nil))
}
