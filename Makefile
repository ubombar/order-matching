all:
	go build -o ./bin/order-matching ./cmd/main.go
	./bin/order-matching