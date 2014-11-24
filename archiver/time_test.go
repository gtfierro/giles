package archiver

import (
	"regexp"
	"testing"
	"time"
)

func TestRelativeTimeRegex(t *testing.T) {
	re := regexp.MustCompile("([-+][0-9]+)(hour|h|minute|min|m|second|sec|s|day|d)")
	var res [][]string
	res = re.FindAllStringSubmatch("-1h", -1)
	if len(res) != 1 {
		t.Error("Found more than 1 match")
	}
	if res[0][1] != "-1" {
		t.Error("Should be -1: ", res[0][1])
	}
	if res[0][2] != "h" {
		t.Error("Should be h: ", res[0][2])
	}

	res = re.FindAllStringSubmatch("+10hour", -1)
	if len(res) != 1 {
		t.Error("Found more than 1 match")
	}
	if res[0][1] != "+10" {
		t.Error("Should be +10: ", res[0][1])
	}
	if res[0][2] != "hour" {
		t.Error("Should be hour: ", res[0][2])
	}

	res = re.FindAllStringSubmatch("+10hour", -1)
	if len(res) != 1 {
		t.Error("Found more than 1 match")
	}
	if res[0][1] != "+10" {
		t.Error("Should be +10: ", res[0][1])
	}
	if res[0][2] != "hour" {
		t.Error("Should be hour: ", res[0][2])
	}
}

func TestParseDuration(t *testing.T) {
	var teststring string
	var err error
	var duration time.Duration
	var shouldbe time.Duration

	teststring = "-1h"
	duration, err = parseIntoDuration(teststring)
	if err != nil {
		t.Error("parsing -1h gave error", err)
	}
	shouldbe = time.Duration(-1) * time.Hour
	if duration != shouldbe {
		t.Error("-1h gave", duration, "but should be", shouldbe)
	}

	teststring = "-1hour"
	duration, err = parseIntoDuration(teststring)
	if err != nil {
		t.Error("parsing -1hour gave error", err)
	}
	shouldbe = time.Duration(-1) * time.Hour
	if duration != shouldbe {
		t.Error("-1hour gave", duration, "but should be", shouldbe)
	}

	teststring = "+1d"
	duration, err = parseIntoDuration(teststring)
	if err != nil {
		t.Error("parsing +1d gave error", err)
	}
	shouldbe = time.Duration(24) * time.Hour
	if duration != shouldbe {
		t.Error("+1d gave", duration, "but should be", shouldbe)
	}

	teststring = "+10days"
	duration, err = parseIntoDuration(teststring)
	if err != nil {
		t.Error("parsing +10days gave error", err)
	}
	shouldbe = time.Duration(240) * time.Hour
	if duration != shouldbe {
		t.Error("+1days gave", duration, "but should be", shouldbe)
	}

	teststring = "-5m"
	duration, err = parseIntoDuration(teststring)
	if err != nil {
		t.Error("parsing -5m gave error", err)
	}
	shouldbe = time.Duration(-5) * time.Minute
	if duration != shouldbe {
		t.Error("-5m gave", duration, "but should be", shouldbe)
	}
}

func TestHandleTime(t *testing.T) {
	var portions []string
	var err error
	var parsedtime time.Time
	var shouldbe time.Time

	shouldbe = time.Now()
	portions = []string{"now", "-1h"}
	parsedtime, err = handleTime(portions)
	if err != nil {
		t.Error(err)
	}
	if !parsedtime.Before(shouldbe) {
		t.Error(parsedtime, "should be before", shouldbe)
	}

	shouldbe = time.Now()
	portions = []string{"now", "+1h", "-5m"}
	parsedtime, err = handleTime(portions)
	if err != nil {
		t.Error(err)
	}
	if !parsedtime.After(shouldbe) {
		t.Error(parsedtime, "should be after", shouldbe)
	}

	shouldbe = time.Now()
	portions = []string{"now", "+1h", "-65m", "+10s"}
	parsedtime, err = handleTime(portions)
	if err != nil {
		t.Error(err)
	}
	if !parsedtime.Before(shouldbe) {
		t.Error(parsedtime, "should be before", shouldbe)
	}
}
