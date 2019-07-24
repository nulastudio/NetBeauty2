OUTPUT=bin
BINARY_LINUX=linux/ncbeauty
BINARY_WIN=win/ncbeauty.exe
BINARY_MAC=osx/ncbeauty
BUILD_FLAGS=-ldflags="-s -w"

build-all: build-win build-linux build-mac

build-win:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(BUILD_FLAGS) -o ./$(OUTPUT)/$(BINARY_WIN) ./src/main/beauty.go

build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o ./$(OUTPUT)/$(BINARY_LINUX) ./src/main/beauty.go

build-mac:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) -o ./$(OUTPUT)/$(BINARY_MAC) ./src/main/beauty.go
