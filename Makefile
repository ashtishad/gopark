run:
	export API_HOST=127.0.0.1 \
	export API_PORT=8080 \
	export DB_USER=postgres \
	export DB_PASSWD=postgres \
	export DB_HOST=127.0.0.1 \
	export DB_PORT=5432 \
	export DB_NAME=gopark \
&& go run main.go
test:
	go test -v ./...
race:
	go run -race .
lint:
	golangci-lint run
