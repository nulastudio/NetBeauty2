OUTPUT=tools
BINARY_WIN_X86=win-x86/ncbeauty.exe
BINARY_WIN_X64=win-x64/ncbeauty.exe
BINARY_LINUX_X64=linux-x64/ncbeauty
BINARY_MAC_X64=osx-x64/ncbeauty
BUILD_FLAGS=-ldflags="-s -w"

build-all: build-win-x86 build-win-x64 build-linux-x64 build-osx-x64

build-win-x86:
	CGO_ENABLED=0 GOOS=windows GOARCH=386 go build $(BUILD_FLAGS) -o ./$(OUTPUT)/$(BINARY_WIN_X86) ./src/main/beauty.go ./src/main/bindata.go

build-win-x64:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(BUILD_FLAGS) -o ./$(OUTPUT)/$(BINARY_WIN_X64) ./src/main/beauty.go ./src/main/bindata.go

build-linux-x64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o ./$(OUTPUT)/$(BINARY_LINUX_X64) ./src/main/beauty.go ./src/main/bindata.go

build-osx-x64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) -o ./$(OUTPUT)/$(BINARY_MAC_X64) ./src/main/beauty.go ./src/main/bindata.go
