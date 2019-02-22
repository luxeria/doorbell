package doorbell

import (
	"errors"
	"github.com/luxeria/doorbell/pkg/rest/auth"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/luxeria/doorbell/pkg/openinghours"
	"github.com/luxeria/doorbell/pkg/ratelimit"
	"github.com/luxeria/doorbell/pkg/rest"
)

type Config struct {
	OpeningHours openinghours.OpeningHours
	RateLimit    *ratelimit.Bucket
	DoorbellCmd  []string
}

type Doorbell struct {
	openingHours openinghours.OpeningHours
	rateLimit    *ratelimit.Bucket
	doorbellCmd  []string
}

func New(c Config) *Doorbell {
	if c.OpeningHours.IsZero() {
		panic("opening hours are invalid (always closed)")
	}

	if c.RateLimit == nil {
		panic("ratelimit bucket is nil")
	}

	if len(c.DoorbellCmd) == 0 {
		panic("doorbell command is empty")
	}

	return &Doorbell{
		openingHours: c.OpeningHours,
		rateLimit:    c.RateLimit,
		doorbellCmd:  c.DoorbellCmd,
	}
}

func (d *Doorbell) Ring() http.Handler {
	return rest.PostRequest(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !d.openingHours.IsOpen() {
			rest.Error(w, r, errors.New("unavailable outside opening hours"), http.StatusServiceUnavailable)
			return
		}

		if !d.rateLimit.Take() {
			rest.Error(w, r, errors.New("rate limit occurred"), http.StatusTooManyRequests)
			return
		}

		cmd := exec.Command(d.doorbellCmd[0], d.doorbellCmd[1:]...)
		err := cmd.Start()
		if err != nil {
			rest.Error(w, r, err, http.StatusInternalServerError)
			return
		}

		if c, ok := auth.ExtractJwtClaims(r); ok {
			log.Printf("%q is ringing doorbell!", c.Subject)
		} else {
			log.Println("unknown user is ringing doorbell!")
		}

		// wait for completion in background
		go func() {
			err := cmd.Wait()
			if err != nil {
				args := strings.Join(d.doorbellCmd, " ")
				if exitErr, ok := err.(*exec.ExitError); ok {
					log.Printf("command `%s` failed (%s): %s", args, exitErr, exitErr.Stderr)
				} else {
					log.Printf("command `%s` failed with error: %s", args, err)
				}
			}
		}()

		rest.JSON(w, "RING", http.StatusOK)
	}))
}