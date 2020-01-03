VERSION=0.1.0
GOLDFLAGS="-X main.version=$(VERSION)"

default: release

clean:
	rm -rf bin

bin:
	mkdir -p bin
	rm -rf bin/*

release: release-windows release-darwin release-linux

release-windows: bin
	GOOS=windows GOARCH=amd64 go build -ldflags=$(GOLDFLAGS) -o bin/got ./cmd/got
	cd bin && tar -cvzf got-v$(VERSION)-windows-amd64.tgz got
	rm bin/got

release-darwin: bin
	GOOS=darwin GOARCH=amd64 go build -ldflags=$(GOLDFLAGS) -o bin/got ./cmd/got
	cd bin && tar -cvzf got-v$(VERSION)-darwin-amd64.tgz got
	rm bin/got

release-linux: bin
	GOOS=linux GOARCH=amd64 go build -ldflags=$(GOLDFLAGS) -o bin/got ./cmd/got
	cd bin && tar -cvzf got-v$(VERSION)-linux-amd64.tgz got
	rm bin/got

.PHONY: bin clean default release
