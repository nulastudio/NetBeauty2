$OUTPUT           = "../Build/tools"
$BINARY_WIN_X86   = "win-x86/nbeauty2.exe"
$BINARY_WIN_X64   = "win-x64/nbeauty2.exe"
$BINARY_LINUX_X64 = "linux-x64/nbeauty2"
$BINARY_MAC_X64   = "osx-x64/nbeauty2"
$BUILD_FLAGS      = '-ldflags="-s -w"'
$PACKAGE          = "github.com/nulastudio/NetBeauty/src/main"

$Env:CGO_ENABLED  = "0"

$Env:GOOS   = "windows"
$Env:GOARCH = "386"
go build ${BUILD_FLAGS} -o ./${OUTPUT}/${BINARY_WIN_X86} $PACKAGE

$Env:GOOS   = "windows"
$Env:GOARCH = "amd64"
go build ${BUILD_FLAGS} -o ./${OUTPUT}/${BINARY_WIN_X64} $PACKAGE

$Env:GOOS   = "linux"
$Env:GOARCH = "amd64"
go build ${BUILD_FLAGS} -o ./${OUTPUT}/${BINARY_LINUX_X64} $PACKAGE

$Env:GOOS   = "darwin"
$Env:GOARCH = "amd64"
go build ${BUILD_FLAGS} -o ./${OUTPUT}/${BINARY_MAC_X64} $PACKAGE
