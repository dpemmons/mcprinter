build:
	go build -o mcprint .

CONFIG_DIR := $(or $(XDG_CONFIG_HOME),$(HOME)/.config)/mcprint

install: build
	mkdir -p ~/.local/bin
	cp mcprint ~/.local/bin/
	mkdir -p $(CONFIG_DIR)
	@test -f $(CONFIG_DIR)/config.env || cp .env.example $(CONFIG_DIR)/config.env

uninstall:
	rm -f ~/.local/bin/mcprint
	@echo "Config left in $(CONFIG_DIR)/config.env — remove manually if desired"

clean:
	rm -f mcprint

test:
	go test ./... -v

.PHONY: build install uninstall clean test
