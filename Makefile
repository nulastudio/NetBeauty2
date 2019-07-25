OUTPUT=tools
BINARY_LINUX_X64=linux-x64/ncbeauty
BINARY_WIN_X64=win-x64/ncbeauty.exe
BINARY_MAC_X64=osx-x64/ncbeauty
BUILD_FLAGS=-ldflags="-s -w"

build-all: build-win-x64 build-linux-x64 build-osx-x64

build-win-x64:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(BUILD_FLAGS) -o ./$(OUTPUT)/$(BINARY_WIN_X64) ./src/main/beauty.go

build-linux-x64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o ./$(OUTPUT)/$(BINARY_LINUX_X64) ./src/main/beauty.go

build-osx-x64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) -o ./$(OUTPUT)/$(BINARY_MAC_X64) ./src/main/beauty.go
