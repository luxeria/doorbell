package openinghours

import (
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	_, err := Parse("Mo 07:00-18:00")
	if err != nil {
		t.Error(err)
	}

	_, err = Parse("We-Fr 22:00-00:00")
	if err != nil {
		t.Error(err)
	}

	_, err = Parse("Mo,Tu-Th,Sa 12:00-20:00")
	if err != nil {
		t.Error(err)
	}

	_, err = Parse("Mo,Tu-Th,Sa-Su 2:00-23:00")
	if err != nil {
		t.Error(err)
	}

	// empty string
	_, err = Parse("")
	if err == nil {
		t.Error("parser did not reject empty string")
	}

	// invalid weekdays
	_, err = Parse(", 00:00-00:00")
	if err == nil {
		t.Error("parser did not reject empty weekday list")
	}

	_, err = Parse("Mo, 00:00-00:00")
	if err == nil {
		t.Error("parser did not reject invalid weekday list")
	}

	_, err = Parse("Mo- 00:00-00:00")
	if err == nil {
		t.Error("parser did not reject invalid weekday range")
	}

	_, err = Parse("-Sa 00:00-00:00")
	if err == nil {
		t.Error("parser did not reject invalid weekday range")
	}

	// invalid time ranges
	_, err = Parse("Mo 00:00-24:00")
	if err == nil {
		t.Error("parser did not reject invalid times")
	}

	_, err = Parse("Mo 07-21")
	if err == nil {
		t.Error("parser did not reject missing minutes")
	}

	_, err = Parse("Mo 00:00")
	if err == nil {
		t.Error("parser did not reject missing range")
	}

	_, err = Parse("Mo 07:00-")
	if err == nil {
		t.Error("parse did not reject missing opening hours")
	}

	_, err = Parse("Mo -23:00")
	if err == nil {
		t.Error("parser did not reject missing closing hours")
	}
}

func assertResult(t *testing.T, openingHours string, datetime string, open bool) {
	o, err := Parse(openingHours)
	if err != nil {
		t.Fatal(err)
	}

	dt, err := time.Parse(time.ANSIC, datetime)
	if err != nil {
		t.Fatal(err)
	}

	if o.IsOpenAt(dt) != open {
		t.Errorf("expected isOpen(%s)=%t for spec '%s'", dt, open, openingHours)
	}
}

func assertOpen(t *testing.T, openingHours string, datetime string) {
	assertResult(t, openingHours, datetime, true)
}

func assertClosed(t *testing.T, openingHours string, datetime string) {
	assertResult(t, openingHours, datetime, false)
}

func TestIsOpen(t *testing.T) {
	assertOpen(t, "Mo 07:00-18:00", "Mon Jan 2 07:00:00 2006")
	assertOpen(t, "Mo 07:00-18:00", "Mon Jan 2 12:34:56 2006")
	assertOpen(t, "Mo 07:00-18:00", "Mon Jan 2 18:00:00 2006")

	assertClosed(t, "Mo 07:00-18:00", "Mon Jan 2 06:59:59 2006")
	assertClosed(t, "Mo 07:00-18:00", "Mon Jan 2 18:00:01 2006")
	assertClosed(t, "Mo 07:00-18:00", "Tue Jan 3 12:34:56 2006")

	assertOpen(t, "Sa-Su 20:00-3:00", "Sat Dec 31 21:00:00 2005")
	assertOpen(t, "Sa-Su 20:00-3:00", "Sun Jan 1 21:00:00 2006")
	assertOpen(t, "Sa-Su 20:00-3:00", "Sun Jan 1 02:00:00 2006")
	assertOpen(t, "Sa-Su 20:00-3:00", "Mon Jan 2 02:00:00 2006")

	assertOpen(t, "Mo 0:00-0:00", "Mon Jan 2 00:00:00 2006")
	assertOpen(t, "Mo 0:00-0:00", "Mon Jan 2 01:00:00 2006")
	assertOpen(t, "Mo 0:00-0:00", "Tue Jan 3 00:00:00 2006")

	assertClosed(t, "Mo 0:00-0:00", "Sun Jan 1 00:00:00 2006")
	assertClosed(t, "Mo 0:00-0:00", "Wed Jan 4 00:00:00 2006")
}
