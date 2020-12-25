$OUTPUT = "Build/tools"
$BINARY_WIN_X86 = "win-x86/ncbeauty.exe"
$BINARY_WIN_X64 = "win-x64/ncbeauty.exe"
$BINARY_LINUX_X64 = "linux-x64/ncbeauty"
$BINARY_MAC_X64 = "osx-x64/ncbeauty"
$BUILD_FLAGS = '-ldflags="-s -w"'

$Package = "github.com/nulastudio/NetCoreBeauty/src/main"

$Env:CGO_ENABLED="0"

$Env:GOOS="windows"
$Env:GOARCH="386"
go build ${BUILD_FLAGS} -o ./${OUTPUT}/${BINARY_WIN_X86} $Package

$Env:GOOS="windows"
$Env:GOARCH="amd64"
go build ${BUILD_FLAGS} -o ./${OUTPUT}/${BINARY_WIN_X64} $Package

$Env:GOOS="linux"
$Env:GOARCH="amd64"
go build ${BUILD_FLAGS} -o ./${OUTPUT}/${BINARY_LINUX_X64} $Package

$Env:GOOS="darwin"
$Env:GOARCH="amd64"
go build ${BUILD_FLAGS} -o ./${OUTPUT}/${BINARY_MAC_X64} $Package
