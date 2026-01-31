build:
	go build -o mcprint .

CONFIG_DIR := $(or $(XDG_CONFIG_HOME),$(HOME)/.config)/mcprint

install: build
	mkdir -p ~/.local/bin
	cp mcprint ~/.local/bin/
	mkdir -p $(CONFIG_DIR)
	@test -f $(CONFIG_DIR)/config.env || cp .env.example $(CONFIG_DIR)/config.env
	@echo "Installed mcprint to ~/.local/bin/"
	@echo "Config at $(CONFIG_DIR)/config.env"
	@case "$$PATH" in *"$$HOME/.local/bin"*) ;; *) \
		echo ""; \
		echo "WARNING: ~/.local/bin is not in your PATH."; \
		echo "Add it by running:"; \
		echo ""; \
		if [ -f "$$HOME/.zshrc" ]; then \
			echo '  echo '\''export PATH="$$HOME/.local/bin:$$PATH"'\'' >> ~/.zshrc'; \
		elif [ -f "$$HOME/.bashrc" ]; then \
			echo '  echo '\''export PATH="$$HOME/.local/bin:$$PATH"'\'' >> ~/.bashrc'; \
		else \
			echo '  echo '\''export PATH="$$HOME/.local/bin:$$PATH"'\'' >> ~/.<your-shell>rc'; \
		fi; \
	esac

uninstall:
	rm -f ~/.local/bin/mcprint
	@echo "Config left in $(CONFIG_DIR)/config.env — remove manually if desired"

clean:
	rm -f mcprint

test:
	go test ./... -v

.PHONY: build install uninstall clean test
