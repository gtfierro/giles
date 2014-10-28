package main

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var supported_formats = []string{"1/2/2006", "1-2-2006", "1/2/2006 04:15", "1-2-2006 04:15", "2006-1-2 15:04:05"}

/**
 * Takes a string specifying a time, and returns a canonical Time object representing that string.
 * To consider: should this instead return the kind of timestamp expected by ReadingDB? Or can that
 * be handled by another method? I think the latter is the way to go on this, that way I can use
 * this method for displaying and the like
 *
 * Need to support the following:
 * now
 * now -1h
 * now +1h -10m
 * %m/%d/%Y
 * %m-%d-%Y
 * %m/%d/%Y %M:%H
 * %m-%d-%Y %M:%H
 * %Y-%m-%dT%H:%M:%S

 Go time layout: Mon Jan 2 15:04:05 -0700 MST 2006
 * 01/02/2006
 * 01-02-2006
 * 01/02/2006 04:15
 * 01-02-2006 04:15
 * 2006-01-02 15:04:05
**/
func handleTime(portions []string) (time.Time, error) {
	ret := time.Now()
	idx := len(portions) - 1
	portions[idx] = strings.Replace(portions[idx], ")", "", -1)
	isNowToken := regexp.MustCompile("now")
	// check if parsing relative timestamps
	if isNowToken.MatchString(portions[0]) {
		for _, val := range portions[1:] {
			// parse the relative duration
			dur, err := parseIntoDuration(val)
			if err != nil {
				return ret, err
			}
			// adjust the time by the duration amount
			ret = ret.Add(dur)
		}
	} else {
		portions[0] = strings.Replace(portions[idx], "(", "", -1)
		portions[0] = strings.Replace(portions[idx], ",", "", -1)
		timestring := strings.Join(portions, " ")
		log.Debug("parsing", timestring)
		for _, format := range supported_formats {
			t, err := time.Parse(format, timestring)
			if err != nil {
				continue
			}
			ret = t
			break
		}
	}
	return ret, nil
}

/**
 * Takes a duration string like -1d, +5minutes, etc and returns a time.Duration object
**/
func parseIntoDuration(str string) (time.Duration, error) {
	var d time.Duration
	/**
	 * important! When editing this regex, make sure that you specify the "or"s as
	 * whole -> subset instead of subset -> whole, that is "second|sec|s" instead of
	 * "s|sec|second". Otherwise, you will find yourself matching "s", but with a tailing
	 * "econd"
	**/
	re := regexp.MustCompile("([-+][0-9]+)(hour|hr|h|minute|min|m|second|sec|s|days|day|d)")
	res := re.FindAllStringSubmatch(str, -1)
	if len(res) != 1 {
		return d, errors.New("Invalid timespec: " + str)
	}

	// handle amount
	i, err := strconv.ParseInt(res[0][1], 10, 64)
	if err != nil {
		return d, err
	}
	d = time.Duration(i)

	// handle units
	switch res[0][2] {
	case "h", "hr", "hour":
		d *= time.Hour
	case "m", "min", "minute":
		d *= time.Minute
	case "s", "sec", "second":
		d *= time.Second
	case "d", "days", "day":
		d *= 24 * time.Hour
	default:
		return d, errors.New("Timespec needs valid units:" + str)
	}

	return d, nil
}
