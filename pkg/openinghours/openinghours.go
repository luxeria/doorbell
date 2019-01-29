package openinghours

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type OpeningHours struct {
	openDays              map[time.Weekday]bool
	opensHour, opensMin   int
	closesHour, closesMin int
}

const daysInWeek = 7

func parseWeekday(s string) (w time.Weekday, err error) {
	switch s {
	case "Mo":
		w = time.Monday
	case "Tu":
		w = time.Tuesday
	case "We":
		w = time.Wednesday
	case "Th":
		w = time.Thursday
	case "Fr":
		w = time.Friday
	case "Sa":
		w = time.Saturday
	case "Su":
		w = time.Sunday
	default:
		err = fmt.Errorf("invalid weekday: %s", s)
	}

	return w, err
}

func parseWeekdayRange(s string, result map[time.Weekday]bool) error {
	days := strings.SplitN(s, "-", 2)
	if len(days) != 2 {
		return fmt.Errorf("invalid weekday range: %s", s)
	}

	start, err := parseWeekday(days[0])
	if err != nil {
		return err
	}

	end, err := parseWeekday(days[1])
	if err != nil {
		return err
	}

	// handle wraparound
	for w := start; ; w = (w + 1) % daysInWeek {
		result[w] = true
		if w == end {
			break
		}
	}

	return nil
}

func parseWeekdayList(list string, result map[time.Weekday]bool) error {
	for _, s := range strings.Split(list, ",") {
		if strings.Contains(s, "-") {
			err := parseWeekdayRange(s, result)
			if err != nil {
				return err
			}
		} else {
			w, err := parseWeekday(s)
			if err != nil {
				return err
			}
			result[w] = true
		}
	}
	return nil
}

func parseTimeRange(s string) (opensHour, opensMin, closesHour, closesMin int, err error) {
	times := strings.SplitN(s, "-", 2)
	if len(times) != 2 {
		err = fmt.Errorf("invalid time range: %s", s)
		return
	}

	const timeLayout = "15:04"
	opens, err := time.Parse(timeLayout, times[0])
	if err != nil {
		return
	}

	closes, err := time.Parse(timeLayout, times[1])
	if err != nil {
		return
	}

	opensHour = opens.Hour()
	opensMin = opens.Minute()
	closesHour = closes.Hour()
	closesMin = closes.Minute()

	return
}

// Parses an opening hour according to https://schema.org/openingHours
func Parse(openingHours string) (o OpeningHours, err error) {
	fields := strings.Fields(openingHours)
	if len(fields) != 2 {
		return o, errors.New("wrong number of components in opening hours")
	}

	// parse weekday constraints
	o.openDays = make(map[time.Weekday]bool)
	err = parseWeekdayList(fields[0], o.openDays)
	if err != nil {
		return o, err
	}

	// parse time constraint
	o.opensHour, o.opensMin, o.closesHour, o.closesMin, err = parseTimeRange(fields[1])
	if err != nil {
		return o, err
	}

	return
}

func (o OpeningHours) IsZero() bool {
	return len(o.openDays) == 0 &&
		o.opensHour == 0 && o.opensMin == 0 &&
		o.closesHour == 0 && o.closesMin == 0
}

func (o *OpeningHours) IsOpen() bool {
	return o.IsOpenAt(time.Now())
}

func (o *OpeningHours) IsOpenAt(t time.Time) bool {
	year, month, day := t.Date()
	opens := time.Date(year, month, day, o.opensHour, o.opensMin, 0, 0, t.Location())
	closes := time.Date(year, month, day, o.closesHour, o.closesMin, 0, 0, t.Location())

	if opens.Equal(t) && closes.Equal(t) {
		// corner case for when `t` is on the edge of a 24 hour shift
		opens = opens.AddDate(0, 0, -1)
		return o.openDays[opens.Weekday()] || o.openDays[closes.Weekday()]
	} else if !opens.Before(closes) {
		// adjust times in case of wraparound
		if opens.After(t) {
			// shift opening hour to the day before
			opens = opens.AddDate(0, 0, -1)
		} else {
			// shift closing hour to the next day
			closes = closes.AddDate(0, 0, 1)
		}
	}

	// check if t is outside opening hours
	if t.Before(opens) || t.After(closes) {
		return false
	}

	// check weekday of when the opening hours started
	return o.openDays[opens.Weekday()]
}
