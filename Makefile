build:
	CGO_ENABLED=0 go build -o menu ./cmd/menu/

run: build
	./menu

test:
	go test ./...

clean:
	rm -f menu

.PHONY: build run test clean
