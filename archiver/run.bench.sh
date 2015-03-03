#!/bin/bash -ex
go test -bench=. -run=X -cpuprofile cpu.out -memprofile mem.out
