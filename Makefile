build:
	go build -o mcprinter .

clean:
	rm -f mcprinter

test:
	go test ./... -v

.PHONY: build clean test
