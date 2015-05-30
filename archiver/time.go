package archiver

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"
)

var unitmultiplier = map[UnitOfTime]uint64{
	UOT_NS: 1000000000,
	UOT_US: 1000000,
	UOT_MS: 1000,
	UOT_S:  1}

func parseReltime(num, units string) (time.Duration, error) {
	var d time.Duration
	i, err := strconv.ParseInt(num, 10, 64)
	if err != nil {
		return d, err
	}
	d = time.Duration(i)
	switch units {
	case "h", "hr", "hour", "hours":
		d *= time.Hour
	case "m", "min", "minute", "minutes":
		d *= time.Minute
	case "s", "sec", "second", "seconds":
		d *= time.Second
	case "us", "usec", "microsecond", "microseconds":
		d *= time.Microsecond
	case "ms", "msec", "millisecond", "milliseconds":
		d *= time.Millisecond
	case "ns", "nsec", "nanosecond", "nanoseconds":
		d *= time.Nanosecond
	case "d", "day", "days":
		d *= 24 * time.Hour
	default:
		err = fmt.Errorf("Invalid unit %v. Must be h,m,s,us,ms,ns,d", units)
	}
	return d, err
}

func parseAbsTime(num, units string) (time.Time, error) {
	var d time.Time
	var err error
	i, err := strconv.ParseUint(num, 10, 64)
	if err != nil {
		return d, err
	}
	uot, err := parseUOT(units)
	if err != nil {
		return d, err
	}
	unixseconds := convertTime(i, uot, UOT_S)
	leftover := i - convertTime(unixseconds, UOT_S, uot)
	unixns := convertTime(leftover, uot, UOT_NS)
	d = time.Unix(int64(unixseconds), int64(unixns))
	return d, err
}

/**
Takes 2 durations and returns the result of them added together
*/
func addDurations(d1, d2 time.Duration) time.Duration {
	d1nano := d1.Nanoseconds()
	d2nano := d2.Nanoseconds()
	res := d1nano + d2nano
	return time.Duration(res) * time.Nanosecond
}

// Takes a duration string like -1d, +5minutes, etc and returns a time.Duration object
func parseIntoDuration(str string) (time.Duration, error) {
	var d time.Duration
	/**
	 * important! When editing this regex, make sure that you specify the "or"s as
	 * whole -> subset instead of subset -> whole, that is "second|sec|s" instead of
	 * "s|sec|second". Otherwise, you will find yourself matching "s", but with a tailing
	 * "econd"
	**/
	re := regexp.MustCompile("([-+]?[0-9]+)(hour|hr|h|minute|min|m|second|sec|s|days|day|d)")
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
	dur, err := parseTimeUnit(res[0][2])
	if err != nil {
		return d, err
	}
	return d * dur, nil
}

func parseTimeUnit(units string) (time.Duration, error) {
	switch units {
	case "h", "hr", "hour", "hours":
		return time.Hour, nil
	case "m", "min", "minute", "minutes":
		return time.Minute, nil
	case "s", "sec", "second", "seconds":
		return time.Second, nil
	case "us", "usec", "microsecond", "microseconds":
		return time.Microsecond, nil
	case "ms", "msec", "millisecond", "milliseconds":
		return time.Millisecond, nil
	case "ns", "nsec", "nanosecond", "nanoseconds":
		return time.Nanosecond, nil
	default:
		return time.Second, fmt.Errorf("Invalid unit %v. Must be h,m,s,us,ms,ns", units)
	}
}

func parseUOT(units string) (UnitOfTime, error) {
	switch units {
	case "s", "sec", "second", "seconds":
		return UOT_S, nil
	case "us", "usec", "microsecond", "microseconds":
		return UOT_US, nil
	case "ms", "msec", "millisecond", "milliseconds":
		return UOT_MS, nil
	case "ns", "nsec", "nanosecond", "nanoseconds":
		return UOT_NS, nil
	default:
		return UOT_S, fmt.Errorf("Invalid unit %v. Must be s,us,ms,ns", units)
	}
}

// Takes a timestamp with accompanying unit of time 'stream_uot' and
// converts it to the unit of time 'target_uot'
func convertTime(time uint64, stream_uot, target_uot UnitOfTime) uint64 {
	if stream_uot == target_uot {
		return time
	}
	if target_uot < stream_uot { // target/stream is > 1, so we can use uint64
		return time * (unitmultiplier[target_uot] / unitmultiplier[stream_uot])
	} else { // if target/stream < 1, we need float64
		return uint64(float64(time) * (float64(unitmultiplier[target_uot]) / float64(unitmultiplier[stream_uot])))
	}
}
