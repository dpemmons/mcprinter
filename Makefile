build:
	go build -o mcprint .

install: build
	mkdir -p ~/.local/bin
	cp mcprint ~/.local/bin/

uninstall:
	rm -f ~/.local/bin/mcprint

clean:
	rm -f mcprint

test:
	go test ./... -v

.PHONY: build install uninstall clean test
