package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	log "github.com/nulastudio/NetBeauty/src/log"
	manager "github.com/nulastudio/NetBeauty/src/manager"
	misc "github.com/nulastudio/NetBeauty/src/misc"
	util "github.com/nulastudio/NetBeauty/src/util"
)

const (
	errorLevel  string = "Error"  // log errors only
	detailLevel string = "Detail" // log useful infos
	infoLevel   string = "Info"   // log everything
)

type depsFileDetail struct {
	deps string
	main string
}

var startupHook = "nbloader"

var workingDir, _ = os.Getwd()

var loglevel string
var beautyDir string
var libsDir = "libraries"
var excludes = ""
var hiddens = ""
var enableDebug = false

func main() {
	misc.Umask()

	initCLI()

	log.LogInfo("running nbeauty...")

	subDirs := make([]string, 0)

	// fix deps.json
	checkedDependencies := []depsFileDetail{}
	dependencies := manager.FindDepsJSON(beautyDir)
	if len(dependencies) != 0 {
		for _, deps := range dependencies {
			deps = strings.ReplaceAll(deps, "\\", "/")
			mainProgram := strings.Replace(path.Base(deps), ".deps.json", "", -1)

			checkedDependencies = append(checkedDependencies, depsFileDetail{
				deps: deps,
				main: mainProgram,
			})
		}

		for _, deps := range checkedDependencies {
			log.LogDetail(fmt.Sprintf("fixing %s", deps.deps))

			success := manager.AddStartUpHookToDeps(deps.deps, startupHook)

			allDeps := manager.FixDeps(deps.deps, deps.main, enableDebug)

			_, _, curSubDirs := moveDeps(allDeps, deps.main)

			subDirs = append(subDirs, curSubDirs...)

			if success {
				log.LogDetail(fmt.Sprintf("%s fixed", deps.deps))
			}
		}
	} else {
		log.LogDetail(fmt.Sprintf("no deps.json found in %s", beautyDir))
		log.LogDetail("skipping")
		os.Exit(0)
	}

	uniqieSubDirs := []string{}
	{
		tmp := map[string]byte{}
		for _, e := range subDirs {
			l := len(tmp)
			tmp[e] = 0
			if len(tmp) != l {
				uniqieSubDirs = append(uniqieSubDirs, e)
			}
		}
	}

	// fix runtimeconfig.json
	runtimeConfigs := manager.FindRuntimeConfigJSON(beautyDir)
	if len(runtimeConfigs) != 0 {
		for _, runtimeConfig := range runtimeConfigs {
			log.LogDetail(fmt.Sprintf("fixing %s", runtimeConfig))

			success := manager.AddStartUpHookToRuntimeConfig(runtimeConfig, startupHook) && manager.FixRuntimeConfig(runtimeConfig, libsDir, uniqieSubDirs)

			if success {
				log.LogDetail(fmt.Sprintf("%s fixed", runtimeConfig))
			}
		}
	} else {
		log.LogDetail(fmt.Sprintf("no runtimeconfig.json found in %s", beautyDir))
		log.LogDetail("skipping")
		os.Exit(0)
	}

	// release nbloader
	log.LogDetail("releasing nbloader.dll")
	if releasePath, err := releaseNBLoader(beautyDir); err != nil {
		log.LogError(fmt.Errorf("release nbloader.dll failed: %s : %s", releasePath, err.Error()), true)
	}

	log.LogDetail("nbeauty done. Enjoy it!")
}

func initCLI() {
	flag.CommandLine = flag.NewFlagSet("nbeauty", flag.ContinueOnError)
	flag.CommandLine.Usage = usage
	flag.CommandLine.SetOutput(os.Stdout)
	flag.StringVar(&loglevel, "loglevel", "Error", `log level. valid values: Error/Detail/Info
Error: Log errors only.
Detail: Log useful infos.
Info: Log everything.
`)
	flag.BoolVar(&enableDebug, "enabledebug", false, `allow 3rd debuggers(like dnSpy) debugs the app`)
	flag.StringVar(&hiddens, "hiddens", "", `dlls that end users never needed, so hide them`)

	flag.Parse()

	args := make([]string, 0)

	// 内置的坑爹flag不对空格做忽略处理
	for _, arg := range flag.Args() {
		if arg != "" && arg != " " {
			args = append(args, arg)
		}
	}
	argv := len(args)

	// 必需参数检查
	if argv == 0 {
		usage()
		os.Exit(0)
	}

	// logLevel检查
	if loglevel != errorLevel && loglevel != detailLevel && loglevel != infoLevel {
		loglevel = errorLevel
	}

	// 设置LogLevel
	log.DefaultLogger.LogLevel = map[string]log.LogLevel{
		errorLevel:  log.Error,
		detailLevel: log.Detail,
		infoLevel:   log.Info,
	}[loglevel]
	manager.Logger.LogLevel = log.DefaultLogger.LogLevel

	switch args[0] {
	default:
		beautyDir = args[0]

		if len(args) >= 2 {
			libsDir = flag.Arg(1)
		}

		if len(args) >= 3 {
			excludes = flag.Arg(2)
		}

		beautyDir = strings.Trim(beautyDir, `"`)
		libsDir = strings.Trim(libsDir, `"`)
		excludes = strings.Trim(excludes, `"`)
		hiddens = strings.Trim(hiddens, `"`)
		absDir, err := filepath.Abs(beautyDir)
		if err != nil {
			log.LogPanic(fmt.Errorf("invalid beautyDir: %s", err.Error()), 1)
		}
		beautyDir = absDir
	}
}

func checkArgumentsCount(excepted int, got int) bool {
	if excepted == got {
		return true
	}
	log.LogPanic(fmt.Errorf("Too few or many arguments, expected %d, got %d", excepted, got), 1)
	return false
}

func usage() {
	fmt.Println("Usage:")
	fmt.Println("nbeauty [--loglevel=(Error|Detail|Info)] [--hiddens=hiddenFiles] <beautyDir> [<libsDir> [<excludes>]]")
	fmt.Println("")
	fmt.Println("Arguments")
	fmt.Println("  <excludes>    dlls that no need to be moved, multi-dlls separated with \";\". Example: dll1.dll;lib*;...")
	fmt.Println("")
	fmt.Println("Options")
	flag.PrintDefaults()
}

func releaseNBLoader(dir string) (string, error) {
	nbloader, err := Asset("nbloader/nbloader.dll")
	loaderPath := dir + "/nbloader.dll"

	if err == nil {
		if err := ioutil.WriteFile(loaderPath, nbloader, 0666); err != nil {
			return loaderPath, err
		}

		return loaderPath, nil
	}

	return loaderPath, err
}

func moveDeps(deps []manager.Deps, entry string) (int, int, []string) {
	var fileMatch = func(file string, sources []string) bool {
		match := false
		for _, pattern := range sources {
			if pattern == "" {
				continue
			}
			if regex, err := regexp.Compile(strings.ReplaceAll(pattern, "*", ".*")); err == nil {
				match = regex.MatchString(file)
				if match {
					break
				}
			}
		}

		return match
	}

	var isContains = func(arr []string, v string) bool {
		for _, c := range arr {
			if c == v {
				return true
			}
		}

		return false
	}

	excludeFiles := strings.Split(excludes, ";")

	realCount, moved, subDirs := 0, 0, make([]string, 0)

	for _, dep := range deps {
		var absDepsFile = ""
		var usingPath = ""
		var exist = false

		for _, filePath := range []string{dep.SecondPath, dep.Path} {
			absDepsFile = path.Join(beautyDir, filePath)
			if util.PathExists(absDepsFile) {
				usingPath = filePath
				exist = true
				break
			}

			if dep.SecondPath == dep.Path {
				break
			}
		}

		if !exist {
			continue
		}

		if fileMatch(dep.Name, excludeFiles) {
			continue
		}

		realCount++

		usingPath2 := strings.ReplaceAll(usingPath, "\\", "/")
		parts := strings.Split(usingPath2, "/")
		subDir := strings.Join(parts[0:len(parts)-1], "/")

		if subDir != "" && !isContains(subDirs, subDir) {
			subDirs = append(subDirs, subDir)
		}

		newAbsDepsFile := path.Join(beautyDir, libsDir, usingPath)
		oldPath := path.Dir(absDepsFile)
		newPath := path.Dir(newAbsDepsFile)

		if !util.EnsureDirExists(newPath, 0777) {
			log.LogError(fmt.Errorf("%s is not writeable", newPath), false)
		}

		if err := os.Rename(absDepsFile, newAbsDepsFile); err == nil {
			moved++
		}

		fileName := strings.TrimSuffix(path.Base(usingPath), path.Ext(usingPath))

		for _, extFile := range []string{".pdb", ".xml"} {
			oldFile := path.Join(oldPath, fileName+extFile)
			newFile := path.Join(newPath, fileName+extFile)
			if util.PathExists(oldFile) {
				os.Rename(oldFile, newFile)
			}
		}

		dir, _ := ioutil.ReadDir(oldPath)

		if len(dir) == 0 {
			os.Remove(oldPath)
		}
	}

	hiddensFiles := strings.Split(hiddens, ";")
	rootFiles := util.GetAllFiles(beautyDir, false)
	for _, rootFile := range rootFiles {
		if fileMatch(rootFile, hiddensFiles) {
			misc.Hide(rootFile)
		}
	}

	return realCount, moved, subDirs
}
