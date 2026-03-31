BINARY  = deadhead
LDFLAGS = -ldflags="-s -w"

.PHONY: build run install clean

build:
	go build $(LDFLAGS) -o $(BINARY) ./cmd/deadhead/

run:
	go run ./cmd/deadhead/ $(ARGS)

install: build
	mv $(BINARY) /usr/local/bin/$(BINARY)

clean:
	rm -f $(BINARY)
