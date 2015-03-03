test:
	go test -v -cpuprofile cpu.out -memprofile mem.out

bench:
	go test -bench=. -run=X -cpuprofile cpu.out -memprofile mem.out

clean:
	rm cpu.out mem.out httphandler.test
