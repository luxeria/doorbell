package recaptcha

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Recaptcha struct {
	secret string
}

type Verification struct {
	Success     bool      `json:"success"`
	Score       float64   `json:"score"`
	Action      string    `json:"action"`
	ChallengeTS time.Time `json:"challenge_ts"`
	Hostname    string    `json:"hostname"`
	ErrorCodes  []string  `json:"error-codes"`
}

func New(secret string) *Recaptcha {
	return &Recaptcha{
		secret: secret,
	}
}

func (r *Recaptcha) Verify(response string) (v Verification, err error) {
	resp, err := http.PostForm("https://www.google.com/recaptcha/api/siteverify",
		url.Values{
			"secret":   {r.secret},
			"response": {response},
		})
	if err != nil {
		return v, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&v)
	if err != nil {
		return v, err
	}

	if len(v.ErrorCodes) > 0 {
		return v, fmt.Errorf("siteverify returned: %s", strings.Join(v.ErrorCodes, ", "))
	}

	if !v.Success {
		return v, errors.New("captcha response token was invalid")
	}

	return v, nil
}