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
	deps       string
	main       string
	fxrVersion string
	rid        string
}

var startupHook = "nbloader"

var workingDir, _ = os.Getwd()

var loglevel string
var beautyDir string
var libsDir = "libraries"
var excludes = ""
var hiddens = ""
var sharedRuntimeMode = false
var enableDebug = false
var usePatch = false
var isNetFx = false

var gitcdn string
var gittree string = ""

func main() {
	misc.Umask()

	manager.EnsureLocalPath()

	initCLI()

	// 设置CDN
	manager.GitCDN = gitcdn
	if gittree != "" {
		manager.GitTree = gittree
	}

	log.LogInfo("running nbeauty...")

	subDirs := make([]string, 0)
	srmMapping := make(map[string]string, 0)

	fxrVersion, rid := "", ""

	useWPF := false

	exeConfig := manager.FindExeConfig(beautyDir)

	if len(exeConfig) != 0 {
		isNetFx = true
	}

	// fix deps.json
	if !isNetFx {
		checkedDependencies := []depsFileDetail{}
		dependencies := manager.FindDepsJSON(beautyDir)
		if len(dependencies) != 0 {
			for _, deps := range dependencies {
				deps = strings.ReplaceAll(deps, "\\", "/")
				mainProgram := strings.Replace(filepath.Base(deps), ".deps.json", "", -1)

				cfxrVersion, crid := manager.FindFXRVersion(deps)

				if fxrVersion == "" || rid == "" {
					fxrVersion, rid = cfxrVersion, crid
				} else if cfxrVersion == fxrVersion || crid == rid {
					log.LogError(fmt.Errorf("Multiple SCD Versions Detected:\n[%s/%s]\n[%s/%s]", fxrVersion, rid, cfxrVersion, crid), true)
				}

				checkedDependencies = append(checkedDependencies, depsFileDetail{
					deps:       deps,
					main:       mainProgram,
					fxrVersion: cfxrVersion,
					rid:        crid,
				})
			}

			// check if pre-build artifact exists
			if fxrVersion != "" && rid != "" {
				// 必须检查
				manager.CheckRunConfigJSON()

				onlineVersion := manager.GetOnlineArtifactsVersion(fxrVersion, rid)
				if usePatch && onlineVersion == "" {
					log.LogError(fmt.Errorf("Artifact does not exist. %s/%s\nYou can report the missing artifact in here: https://github.com/nulastudio/NetBeauty2/discussions/36", fxrVersion, rid), true)
				}
			}

			for _, deps := range checkedDependencies {
				log.LogDetail(fmt.Sprintf("fixing %s", deps.deps))

				SCDMode := deps.fxrVersion != "" && deps.rid != ""

				if SCDMode {
					log.LogDetail("SCD Mode: Yes")
					log.LogDetail(fmt.Sprintf("SCD Version: %s, %s", deps.fxrVersion, deps.rid))

					if usePatch {
						log.LogDetail("Use Patch: Yes")
					} else {
						log.LogDetail("Use Patch: No")
					}
				} else {
					log.LogDetail("SCD Mode: No")
					log.LogDetail("Use Patch: No")
				}

				success := manager.AddStartUpHookToDeps(deps.deps, startupHook)

				usePatch = SCDMode && usePatch

				allDeps, _useWPF, _ := manager.FixDeps(deps.deps, deps.main, enableDebug, usePatch, sharedRuntimeMode)

				useWPF = _useWPF

				if sharedRuntimeMode {
					log.LogDetail("Shared Runtime Mode: Yes")
					log.LogDetail("moving deps may take some time")
				} else {
					log.LogDetail("Shared Runtime Mode: No")
				}

				_, _, curSubDirs, _srmMapping := moveDeps(allDeps, deps.main, sharedRuntimeMode)

				srmMapping = _srmMapping
				subDirs = append(subDirs, curSubDirs...)

				if success {
					log.LogDetail(fmt.Sprintf("%s fixed", deps.deps))
				}
			}

			// patch
			if usePatch && fxrVersion != "" && rid != "" {
				patch(fxrVersion, rid)
			}
		} else {
			log.LogDetail(fmt.Sprintf("no deps.json found in %s", beautyDir))
			log.LogDetail("skipping")
			os.Exit(0)
		}
	} else {
		for _, appConfig := range exeConfig {
			appConfig = strings.ReplaceAll(appConfig, "\\", "/")
			mainProgram := strings.Replace(filepath.Base(appConfig), ".exe.config", "", -1)

			log.LogDetail(fmt.Sprintf("fixing %s", appConfig))

			log.LogDetail(".Net Fx: Yes")

			allDeps, success := manager.FixExeConfig(appConfig, libsDir)

			moveDeps(allDeps, mainProgram, false)

			if success {
				log.LogDetail(fmt.Sprintf("%s fixed", appConfig))
			}
		}
	}

	uniqieSubDirs := []string{}
	if !isNetFx {
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
	if !isNetFx {
		runtimeConfigs := manager.FindRuntimeConfigJSON(beautyDir)
		if len(runtimeConfigs) != 0 {
			for _, runtimeConfig := range runtimeConfigs {
				log.LogDetail(fmt.Sprintf("fixing %s", runtimeConfig))

				success := manager.AddStartUpHookToRuntimeConfig(runtimeConfig, startupHook) && manager.FixRuntimeConfig(runtimeConfig, libsDir, uniqieSubDirs, srmMapping, sharedRuntimeMode, usePatch, useWPF)

				if success {
					log.LogDetail(fmt.Sprintf("%s fixed", runtimeConfig))
				}
			}
		} else {
			log.LogDetail(fmt.Sprintf("no runtimeconfig.json found in %s", beautyDir))
			log.LogDetail("skipping")
			os.Exit(0)
		}
	}

	// release nbloader
	if !isNetFx {
		var loaderDir = beautyDir
		if usePatch {
			loaderDir = filepath.Join(beautyDir, libsDir)
		}
		log.LogDetail("releasing nbloader.dll")
		if releasePath, err := releaseNBLoader(loaderDir); err != nil {
			log.LogError(fmt.Errorf("release nbloader.dll failed: %s : %s", releasePath, err.Error()), true)
		}
	}

	// hide files
	hideFiles()

	log.LogDetail("nbeauty done. Enjoy it!")
}

func initCLI() {
	flag.CommandLine = flag.NewFlagSet("nbeauty", flag.ContinueOnError)
	flag.CommandLine.Usage = usage
	flag.CommandLine.SetOutput(os.Stdout)
	flag.StringVar(&gitcdn, "gitcdn", "", `[.NET Core App Only] specify a HostFXRPatcher mirror repo if you have troble in connecting github.
RECOMMEND https://gitee.com/liesauer/HostFXRPatcher for mainland china users.
`)
	flag.StringVar(&gittree, "gittree", "", `[.NET Core App Only] specify to a valid git branch or any bits commit hash(up to 40) to grab the specific artifacts and won't get updates any more.
default is master, means that you always use the latest artifacts.
NOTE: please provide as longer commit hash as you can, otherwise it may can not be determined as a valid unique commit hash.
`)
	flag.StringVar(&loglevel, "loglevel", "Error", `log level. valid values: Error/Detail/Info
Error: Log errors only.
Detail: Log useful infos.
Info: Log everything.
`)
	flag.BoolVar(&sharedRuntimeMode, "srmode", false, `[.NET Core App Only] share the runtime between apps`)
	flag.BoolVar(&enableDebug, "enabledebug", false, `[.NET Core App Only] allow 3rd debuggers(like dnSpy) debugs the app`)
	flag.BoolVar(&usePatch, "usepatch", false, `[.NET Core App Only] use the patched hostfxr to reduce files`)
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

func exit() {
	os.Exit(0)
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
	log.LogInfo(fmt.Sprintf("backuping fxr to %s", absFxrBakName))

	if _, err := util.CopyFile(absFxrName, absFxrBakName); err != nil {
		log.LogError(fmt.Errorf("backup failed: %s", err.Error()), false)
		return false
	}

	success := manager.CopyArtifactTo(fxrVersion, rid, beautyDir)
	if success {
		log.LogInfo("patch succeeded")
	} else {
		fmt.Println("patch failed")
	}

	return success
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

func fileMatch(file string, sources []string) bool {
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

func moveDeps(deps []manager.Deps, entry string, sharedRuntimeMode bool) (int, int, []string, map[string]string) {
	var isContains = func(arr []string, v string) bool {
		for _, c := range arr {
			if c == v {
				return true
			}
		}

		return false
	}

	excludeFiles := strings.Split(excludes, ";")

	realCount, moved, subDirs, srmMapping := 0, 0, make([]string, 0), make(map[string]string, 0)

	for _, dep := range deps {
		var absDepsFile = ""
		var usingPath = ""
		var exist = false

		for _, filePath := range []string{dep.SecondPath, dep.Path} {
			absDepsFile = filepath.Join(beautyDir, filePath)
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

		if !isNetFx {
			/**
			* !usePatch + !enableDebug = !move +  delete
			* !usePatch +  enableDebug = !move + !delete
			*  usePatch + !enableDebug = !move +  delete
			*  usePatch +  enableDebug =  move + !delete
			 */
			if strings.Contains(dep.Name, "mscordaccore") ||
				strings.Contains(dep.Name, "mscordbi") {
				if !enableDebug {
					os.Remove(absDepsFile)
					continue
				} else if !usePatch {
					continue
				}
			}
		}

		realCount++

		usingPath2 := strings.ReplaceAll(usingPath, "\\", "/")
		parts := strings.Split(usingPath2, "/")
		fileName := parts[len(parts)-1]
		subDir := strings.Join(parts[0:len(parts)-1], "/")

		if dep.Type != manager.Resource && subDir != "" && !isContains(subDirs, subDir) {
			subDirs = append(subDirs, subDir)
		}

		// native不能使用分层结构（多层依赖会导致加载不了dll）
		if !isNetFx && sharedRuntimeMode {
			if dep.Type != manager.Native {
				md5, _ := util.GetFileMD5(absDepsFile)
				if md5 == "" {
					md5 = "generic"
				}
				parts = append(parts, md5, fileName)
				srmKey := fileName
				if dep.Type == manager.Resource {
					srmKey = parts[0] + "/" + srmKey
				}
				srmMapping[srmKey] = md5
				usingPath = strings.Join(parts, "/")
			} else {
				appID, _ := util.GetStringMD5(entry)
				parts = append([]string{"srm_native", appID}, parts...)
				usingPath = strings.Join(parts, "/")
			}
		}

		if !isNetFx && dep.Type == manager.Resource {
			parts = append([]string{"locales"}, parts...)
			usingPath = strings.Join(parts, "/")
		}

		newAbsDepsFile, _ := filepath.Abs(beautyDir + "/" + libsDir + "/" + usingPath)
		oldPath := filepath.Dir(absDepsFile)
		newPath := filepath.Dir(newAbsDepsFile)

		if !util.EnsureDirExists(newPath, 0777) {
			log.LogError(fmt.Errorf("%s is not writeable", newPath), false)
		}

		if err := os.Rename(absDepsFile, newAbsDepsFile); err == nil {
			moved++
		} else {
			fmt.Println(err.Error())
		}

		for _, extFile := range []string{".pdb", ".xml"} {
			oldFile := filepath.Join(oldPath, fileName+extFile)
			newFile := filepath.Join(newPath, fileName+extFile)
			if util.PathExists(oldFile) {
				os.Rename(oldFile, newFile)
			}
		}

		dir, _ := ioutil.ReadDir(oldPath)

		if len(dir) == 0 {
			os.Remove(oldPath)
		}
	}

	return realCount, moved, subDirs, srmMapping
}

func hideFiles() {
	hiddensFiles := strings.Split(hiddens, ";")
	rootFiles := util.GetAllFiles(beautyDir, false)
	for _, rootFile := range rootFiles {
		if fileMatch(rootFile, hiddensFiles) {
			if err := misc.Hide(rootFile); err != nil {
				log.LogError(fmt.Errorf("hide file failed: %s : %s", rootFile, err.Error()), false)
			}
		}
	}
}
