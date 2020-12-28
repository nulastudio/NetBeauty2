package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	log "github.com/nulastudio/NetCoreBeauty/src/log"
	manager "github.com/nulastudio/NetCoreBeauty/src/manager"
	util "github.com/nulastudio/NetCoreBeauty/src/util"

	"github.com/bitly/go-simplejson"
)

const (
	errorLevel  string = "Error"  // log errors only
	detailLevel string = "Detail" // log useful infos
	infoLevel   string = "Info"   // log everything
)

var workingDir, _ = os.Getwd()
var runtimeCompatibilityJSON *simplejson.Json
var runtimeSupportedJSON *simplejson.Json

var gitcdn string
var gittree string = ""
var nopatch bool
var noflag bool
var force bool
var loglevel string
var beautyDir string
var libsDir = "libraries"
var excludes = ""
var hiddens = ""

func main() {
	Umask()

	manager.EnsureLocalPath()

	initCLI()

	// 设置CDN
	manager.GitCDN = gitcdn
	if gittree != "" {
		manager.GitTree = gittree
	}

	log.LogInfo("running ncbeauty...")

	beautyCheck := path.Join(beautyDir, "NetCoreBeauty")

	// 检查是否已beauty
	if util.PathExists(beautyCheck) {
		if !force {
			log.LogDetail("already beauty. Enjoy it!")
			return
		}
		log.LogDetail("already beauty but you are in force mode, continuing...")
	}

	// 必须检查
	manager.CheckRunConfigJSON()

	// fix runtimeconfig.json
	runtimeConfigs := manager.FindRuntimeConfigJSON(beautyDir)
	if len(runtimeConfigs) != 0 {
		for _, runtimeConfig := range runtimeConfigs {
			log.LogDetail(fmt.Sprintf("fixing %s", runtimeConfig))
			manager.FixRuntimeConfig(runtimeConfig, libsDir)
			log.LogDetail(fmt.Sprintf("%s fixed", runtimeConfig))
		}
	} else {
		log.LogDetail(fmt.Sprintf("no runtimeconfig.json found in %s", beautyDir))
		log.LogDetail("skipping")
		os.Exit(0)
	}

	// fix deps.json
	dependencies := manager.FindDepsJSON(beautyDir)
	if len(dependencies) != 0 {
		for _, deps := range dependencies {
			log.LogDetail(fmt.Sprintf("fixing %s", deps))
			deps = strings.ReplaceAll(deps, "\\", "/")
			mainProgram := strings.Replace(path.Base(deps), ".deps.json", "", -1)
			depsFiles, fxrVersion, rid := manager.FixDeps(deps)
			// patch
			if nopatch {
				fmt.Println("hostfxr patch has been disable, skipped")
			} else if fxrVersion != "" && rid != "" {
				patch(fxrVersion, rid)
			} else {
				log.LogError(errors.New("incomplete fxr info, skipping patch"), false)
			}
			if len(depsFiles) == 0 {
				continue
			}
			log.LogDetail(fmt.Sprintf("%s fixed", deps))
			log.LogInfo("moving runtime...")
			moved := moveDeps(depsFiles, mainProgram)
			log.LogDetail(fmt.Sprintf("%d of %d runtime files moved", moved, len(depsFiles)))
		}
	} else {
		log.LogDetail(fmt.Sprintf("no deps.json found in %s", beautyDir))
		log.LogDetail("skipping")
		os.Exit(0)
	}

	// 写入beauty标记
	if !noflag {
		if err := ioutil.WriteFile(beautyCheck, nil, 0666); err != nil {
			log.LogPanic(fmt.Errorf("beauty sign failed: %s", err.Error()), 1)
		}
	}

	log.LogDetail("ncbeauty done. Enjoy it!")
}

func initCLI() {
	flag.CommandLine = flag.NewFlagSet("ncbeauty", flag.ContinueOnError)
	flag.CommandLine.Usage = usage
	flag.CommandLine.SetOutput(os.Stdout)
	flag.StringVar(&gitcdn, "gitcdn", "", `specify a HostFXRPatcher mirror repo if you have troble in connecting github.
RECOMMEND https://gitee.com/liesauer/HostFXRPatcher for mainland china users.
`)
	flag.StringVar(&gittree, "gittree", "", `specify to a valid git branch or any bits commit hash(up to 40) to grab the specific artifacts and won't get updates any more.
default is master, means that you always use the latest artifacts.
NOTE: please provide as longer commit hash as you can, otherwise it may can not be determined as a valid unique commit hash.
`)
	flag.StringVar(&loglevel, "loglevel", "Error", `log level. valid values: Error/Detail/Info
Error: Log errors only.
Detail: Log useful infos.
Info: Log everything.
`)
	flag.BoolVar(&nopatch, "nopatch", false, `disable hostfxr patch.
DO NOT DISABLE!!!
hostfxr patch fixes https://github.com/nulastudio/NetCoreBeauty/issues/1`)
	flag.BoolVar(&force, "force", false, `disable beauty checking and force beauty again.`)
	flag.BoolVar(&noflag, "noflag", false, `do not generate NetCoreBeauty flag file.`)
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
	case "setcdn":
		checkArgumentsCount(2, argv)
		if manager.SetCDN(strings.Trim(args[1], `"`)) {
			fmt.Println("set default git cdn successfully")
		} else {
			fmt.Println("set default git cdn failed")
		}
		exit()
	case "getcdn":
		checkArgumentsCount(1, argv)
		cdn := manager.GetCDN()
		if cdn == "" {
			fmt.Println("default git cdn has not been set yet")
		} else {
			fmt.Printf("current default git cdn: %s\n", cdn)
		}
		exit()
	case "delcdn":
		checkArgumentsCount(1, argv)
		cdn := manager.GetCDN()
		if cdn == "" {
			fmt.Println("default git cdn has not been set yet")
		} else {
			manager.DelCDN()
			fmt.Printf("current default git cdn has been deleted, it was: [%s] before\n", cdn)
		}
		exit()
	default:
		if gitcdn == "" {
			cdn := manager.GetCDN()
			if cdn == "" {
				gitcdn = "https://github.com/nulastudio/HostFXRPatcher"
			} else {
				gitcdn = cdn
			}
		}

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
	fmt.Println("ncbeauty [--force=(True|False)] [--gitcdn=<gitcdn>] [--gittree=<gittree>] [--loglevel=(Error|Detail|Info)] [--nopatch=(True|False)] [--noflag=(True|False)] [--hiddens=hiddenFiles] <beautyDir> [<libsDir> [<excludes>]]")
	fmt.Println("")
	fmt.Println("Arguments")
	fmt.Println("  <excludes>    dlls that no need to be moved, multi-dlls separated with \";\". Example: dll1.dll;lib*;...")
	fmt.Println("")
	fmt.Println("Options")
	flag.PrintDefaults()
	fmt.Println("")
	fmt.Println("Setting GitCDN")
	fmt.Println("ncbeauty [--loglevel=(Error|Detail|Info)] setcdn <gitcdn>")
	fmt.Println("  set current default git cdn, can be override by --gitcdn.")
	fmt.Println("ncbeauty [--loglevel=(Error|Detail|Info)] getcdn")
	fmt.Println("  print current default git cdn.")
	fmt.Println("ncbeauty [--loglevel=(Error|Detail|Info)] delcdn")
	fmt.Println("  remove current default git cdn, after removed, use --gitcdn to specify.")
}

func exit() {
	os.Exit(0)
}

func patch(fxrVersion string, rid string) bool {
	log.LogDetail("patching hostfxr...")

	crid := manager.FindCompatibleRID(rid)
	fxrName := manager.GetHostFXRNameByRID(rid)
	if crid == "" {
		log.LogPanic(fmt.Errorf("cannot find a compatible rid for %s", rid), 1)
	}

	log.LogDetail(fmt.Sprintf("using compatible rid %s for %s", crid, rid))
	rid = crid

	localVersion := manager.GetLocalArtifactsVersion(fxrVersion, rid)
	onlineVersion := manager.GetOnlineArtifactsVersion(fxrVersion, rid)
	if localVersion != onlineVersion {
		log.LogDetail(fmt.Sprintf("downloading patched hostfxr: %s/%s", fxrVersion, rid))

		if !manager.DownloadArtifact(fxrVersion, rid) || !manager.WriteLocalArtifactsVersion(fxrVersion, rid, onlineVersion) {
			log.LogPanic(errors.New("download patch failed"), 1)
		}
	}

	absFxrName := path.Join(beautyDir, fxrName)
	absFxrBakName := absFxrName + ".bak"
	log.LogInfo(fmt.Sprintf("backuping fxr to %s\n", absFxrBakName))

	if util.PathExists(absFxrBakName) && !force {
		log.LogDetail("fxr backup found, skipped")
	} else {
		if _, err := util.CopyFile(absFxrName, absFxrBakName); err != nil {
			log.LogError(fmt.Errorf("backup failed: %s", err.Error()), false)
			return false
		}
	}

	success := manager.CopyArtifactTo(fxrVersion, rid, beautyDir)
	if success {
		log.LogInfo("patch succeeded")
	} else {
		fmt.Println("patch failed")
	}

	return success
}

func moveDeps(depsFiles []string, mainProgram string) int {
	excludeFiles := strings.Split(excludes, ";")
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
	moved := 0
	for _, depsFile := range depsFiles {
		if strings.Join([]string{mainProgram, "dll"}, ".") == depsFile ||
			strings.Contains(depsFile, "hostfxr") ||
			strings.Contains(depsFile, "hostpolicy") ||
			fileMatch(depsFile, excludeFiles) {
			// NOTE: 计数加一，不然每次看到日志的文件移动数少3会造成疑惑
			moved++
			continue
		}

		absDepsFile := path.Join(beautyDir, depsFile)
		absDesFile := path.Join(beautyDir, libsDir, depsFile)
		oldPath := path.Dir(absDepsFile)
		newPath := path.Dir(absDesFile)
		if util.PathExists(absDepsFile) {
			if !util.EnsureDirExists(newPath, 0777) {
				log.LogError(fmt.Errorf("%s is not writeable", newPath), false)
			}
			if err := os.Rename(absDepsFile, absDesFile); err == nil {
				moved++
			}

			// NOTE: pdb、xml跟随程序集
			// TODO: 提供一个选项，自由选择xml：跟随主程序、跟随程序集、跟随两者
			fileName := strings.TrimSuffix(path.Base(depsFile), path.Ext(depsFile))
			extFiles := []string{".pdb", ".xml"}
			for _, extFile := range extFiles {
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
	}

	hiddensFiles := strings.Split(hiddens, ";")
	rootFiles := manager.GetAllFiles(beautyDir, false)
	for _, rootFile := range rootFiles {
		if fileMatch(rootFile, hiddensFiles) {
			Hide(rootFile)
		}
	}

	return moved
}
