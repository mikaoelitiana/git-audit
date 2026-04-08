BINARY  := git-audit
CMD     := ./cmd/git-audit
VERSION := 1.0.0

.PHONY: build install run tidy clean

build:
	go build -ldflags="-s -w" -o $(BINARY) $(CMD)

install:
	go install $(CMD)

run:
	go run $(CMD) .

tidy:
	go mod tidy

clean:
	rm -f $(BINARY)

# Cross-compile targets
build-linux:
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(BINARY)-linux-amd64 $(CMD)

build-mac:
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o $(BINARY)-darwin-arm64 $(CMD)

build-windows:
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $(BINARY)-windows-amd64.exe $(CMD)
