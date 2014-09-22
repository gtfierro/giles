package main

type TSDB interface {
	Add(*SmapReading) bool
	Prev([]string, uint64, uint32) ([]SmapResponse, error)
	Next([]string, uint64, uint32) ([]SmapResponse, error)
	GetData([]string, uint64, uint64) ([]SmapResponse, error)
	Connect()
	DoWrites()
}
