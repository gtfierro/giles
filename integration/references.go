package main

import (
	"code.google.com/p/go-uuid/uuid"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var timeFinder = regexp.MustCompile(`\$TIME_([NMU]?S)\(([0-9]+)\)`)
var uuidFinder = regexp.MustCompile(`\$UUID\(([0-9]+)\)`)

type Manager struct {
	UUIDS map[int]string
	Times map[int]int64
}

func NewManager() *Manager {
	return &Manager{
		UUIDS: make(map[int]string),
		Times: make(map[int]int64),
	}
}

// Retrieves the UUID with the given id. This is created
// if it does not already exist
func (m *Manager) GetUUID(id int) string {
	if retUuid, found := m.UUIDS[id]; found {
		return retUuid
	}
	newUUID := uuid.New()
	m.UUIDS[id] = newUUID
	return newUUID
}

// Retrieves the Time with the given id. This is created
// if it does not already exist (nanoseconds)
func (m *Manager) GetTime(id int) int64 {
	if retTime, found := m.Times[id]; found {
		return retTime
	}
	newTime := int64(time.Now().UnixNano())
	m.Times[id] = newTime
	return newTime
}

func (m *Manager) ParseData(data string) string {
	if timeFinder.MatchString(data) {
		found := timeFinder.FindAllStringSubmatch(data, -1)
		for _, match := range found {
			id, _ := strconv.ParseInt(match[2], 10, 0)
			time.Sleep(time.Duration(convertTime(1e12, match[1])) * time.Nanosecond)
			t := convertTime(m.GetTime(int(id)), match[1])
			ts := fmt.Sprintf("%v", t)
			data = strings.Replace(data, match[0], ts, 1)
		}
	}
	if uuidFinder.MatchString(data) {
		found := uuidFinder.FindAllStringSubmatch(data, -1)
		for _, match := range found {
			id, _ := strconv.ParseInt(match[1], 10, 0)
			useUuid := m.GetUUID(int(id))
			data = strings.Replace(data, match[0], useUuid, 1)
		}
	}
	return data
}
