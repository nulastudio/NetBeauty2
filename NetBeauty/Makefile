OUTPUT           = ../Build/tools
BINARY_WIN_X86   = win-x86/nbeauty2.exe
BINARY_WIN_X64   = win-x64/nbeauty2.exe
BINARY_LINUX_X64 = linux-x64/nbeauty2
BINARY_MAC_X64   = osx-x64/nbeauty2
BINARY_MAC_ARM64 = osx-arm64/nbeauty2
BUILD_FLAGS      = -ldflags="-s -w"
PACKAGE          = "github.com/nulastudio/NetBeauty/src/main"

build-all: build-win-x86 build-win-x64 build-linux-x64 build-osx-x64 build-osx-arm64

build-win-x86:
	CGO_ENABLED=0 GOOS=windows GOARCH=386 go build $(BUILD_FLAGS) -o ./$(OUTPUT)/$(BINARY_WIN_X86) $(PACKAGE)

build-win-x64:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build $(BUILD_FLAGS) -o ./$(OUTPUT)/$(BINARY_WIN_X64) $(PACKAGE)

build-linux-x64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o ./$(OUTPUT)/$(BINARY_LINUX_X64) $(PACKAGE)

build-osx-x64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) -o ./$(OUTPUT)/$(BINARY_MAC_X64) $(PACKAGE)

build-osx-arm64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build $(BUILD_FLAGS) -o ./$(OUTPUT)/$(BINARY_MAC_ARM64) $(PACKAGE)
