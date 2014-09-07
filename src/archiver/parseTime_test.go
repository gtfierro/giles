package main

import (
	"regexp"
	"testing"
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
