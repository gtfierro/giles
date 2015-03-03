#!/bin/bash -ex
go test -v -cpuprofile cpu.out -memprofile mem.out
