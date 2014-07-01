# Benchmarking

`ab -c 100 -n 50000 -e out -r -p data.json -T "application/json" http://127.0.0.1:8079/add`

`go run bench.go`
