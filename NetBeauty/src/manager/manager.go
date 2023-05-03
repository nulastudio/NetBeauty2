package manager

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/beevik/etree"
	"github.com/bitly/go-simplejson"

	log "github.com/nulastudio/NetBeauty/src/log"
	"github.com/nulastudio/NetBeauty/src/util"
)

type DepsType int

const (
	Resource DepsType = iota
	Assembly
	Native
)

type analyzedDeps struct {
	Category   map[string]interface{}
	ItemKey    string
	Name       string
	Path       string
	SecondPath string
	Type       DepsType
	Locale     string
}

type Deps struct {
	Name       string
	Path       string
	SecondPath string
	Type       DepsType
	Locale     string
}

// GitCDN git仓库镜像（默认为github）
var GitCDN = "https://github.com/nulastudio/HostFXRPatcher"

// GitTree git仓库分支（默认为master），支持任意有效分支名、任意长度commit hash（最高40位，为了保证commit hash唯一性，请尽可能提供更长的commit hash，否则将可能无法被识别）
var GitTree = "master"

// Logger 日志记录器
var Logger = log.DefaultLogger

var timeout = 60 * time.Second

var localPath = filepath.Clean(os.TempDir()) + "/NetCoreBeauty"
var localArtifactsPath = localPath + "/artifacts"
var artifactsVersionTXT = "/ArtifactsVersion.txt"
var gitCDNTXT = "/git.cdn"
var artifactsVersionJSON = "/ArtifactsVersion.json"
var onlineArtifactsVersionJSON = "/OnlineArtifactsVersion.json"
var artifactsVersionOldPath = localArtifactsPath + artifactsVersionTXT
var gitCDNPath = localPath + gitCDNTXT
var artifactsVersionPath = localArtifactsPath + artifactsVersionJSON
var onlineArtifactsVersionPath = localArtifactsPath + onlineArtifactsVersionJSON

var runtimeCompatibilityJSONName = "runtime.compatibility.json"
var runtimeSupportedJSONName = "runtime.supported.json"

var pathNotWriteableErr = "cannot create path or path is not writeable: %s"
var getLocalArtifactsVersionErr = "get local artifacts version failed: %s"
var encodeJSONErr = "cannot encode json: %s"

var onlineVersionCache *simplejson.Json = nil

// EnsureLocalPath 确保本地目录存在
func EnsureLocalPath() bool {
	return util.EnsureDirExists(localArtifactsPath, 0777)
}

func formatError(format string, err error) string {
	return fmt.Sprintf(format, err)
}

// FindRuntimeConfigJSON 寻找指定目录下的*runtimeconfig*.json
func FindRuntimeConfigJSON(dir string) []string {
	files, err := filepath.Glob(filepath.Join(dir, "*runtimeconfig*.json"))
	if err != nil {
		log.LogDetail(formatError("find runtimeconfig.json failed: %s", err))
	}
	return files
}

// FindExeConfig 寻找指定目录下的*exe.config
func FindExeConfig(dir string) []string {
	files, err := filepath.Glob(path.Join(dir, "*exe.config"))
	if err != nil {
		log.LogDetail(formatError("find exe.config failed: %s", err))
	}
	return files
}

// FindDepsJSON 寻找指定目录下的*deps.json
func FindDepsJSON(dir string) []string {
	files, err := filepath.Glob(path.Join(dir, "*deps.json"))
	if err != nil {
		log.LogDetail(formatError("find deps.json failed: %s", err))
	}
	return files
}

// AddStartUpHookToDeps 添加nbloader启动时钩子到deps.json
func AddStartUpHookToDeps(deps string, hook string) bool {
	jsonBytes, err := ioutil.ReadFile(deps)
	if err != nil {
		log.LogError(fmt.Errorf("can not read deps.json: %s : %s", deps, err.Error()), false)
		return false
	}

	json, err := simplejson.NewJson(jsonBytes)
	if err != nil {
		log.LogError(fmt.Errorf("invalid deps.json: %s : %s", deps, err.Error()), false)
		return false
	}

	runtimeTarget, _ := json.GetPath("runtimeTarget", "name").String()

	json.SetPath([]string{
		"targets",
		runtimeTarget,
		hook,
		"runtime",
		hook + ".dll",
	}, make(map[string]interface{}))

	json.SetPath([]string{
		"libraries",
		hook,
	}, map[string]interface{}{
		"type":        "project",
		"serviceable": false,
		"sha512":      "",
	})

	jsonBytes, _ = json.EncodePretty()
	if err := ioutil.WriteFile(deps, jsonBytes, 0666); err != nil {
		log.LogError(fmt.Errorf("add startup hook to deps.json failed: %s : %s", deps, err.Error()), false)
		return false
	}

	return true
}

// AddStartUpHookToRuntimeConfig 添加nbloader启动时钩子到runtimeconfig.json
func AddStartUpHookToRuntimeConfig(runtimeConfig string, hook string) bool {
	jsonBytes, err := ioutil.ReadFile(runtimeConfig)
	if err != nil {
		log.LogError(fmt.Errorf("can not read runtimeconfig.json: %s : %s", runtimeConfig, err.Error()), false)
		return false
	}

	json, err := simplejson.NewJson(jsonBytes)
	if err != nil {
		log.LogError(fmt.Errorf("invalid runtimeconfig.json: %s : %s", runtimeConfig, err.Error()), false)
		return false
	}

	json.SetPath([]string{
		"runtimeOptions",
		"configProperties",
		"STARTUP_HOOKS",
	}, hook)

	jsonBytes, _ = json.EncodePretty()
	if err := ioutil.WriteFile(runtimeConfig, jsonBytes, 0666); err != nil {
		log.LogError(fmt.Errorf("add startup hook to runtimeconfig.json failed: %s : %s", runtimeConfig, err.Error()), false)
		return false
	}

	return true
}

// FixExeConfig 添加libs到exe.config
func FixExeConfig(exeConfig string, libsDir string) ([]Deps, bool) {
	var allDeps = make([]Deps, 0)

	doc := etree.NewDocument()
	if err := doc.ReadFromFile(exeConfig); err != nil {
		log.LogError(fmt.Errorf("can not read exe.config: %s : %s", exeConfig, err.Error()), false)
		return allDeps, false
	}

	assemblyBindings := doc.FindElements("./configuration/runtime/assemblyBinding")

	if len(assemblyBindings) == 0 {
		return allDeps, true
	}

	for i, assemblyBinding := range assemblyBindings {
		assemblyIdentity := assemblyBinding.FindElement("./dependentAssembly/assemblyIdentity")

		if assemblyIdentity == nil {
			continue
		}

		dllName := assemblyIdentity.SelectAttrValue("name", "")

		if dllName == "" {
			continue
		}

		dllName += ".dll"

		allDeps = append(allDeps, Deps{
			Name:       dllName,
			Path:       dllName,
			SecondPath: dllName,
			Type:       Assembly,
			Locale:     "",
		})

		if i == 0 {
			probing := assemblyBinding.SelectElement("probing")

			if probing == nil {
				probing = assemblyBinding.CreateElement("probing")
			}

			privatePaths := make([]string, 0)

			privatePathStr := probing.SelectAttrValue("privatePath", "")

			if privatePathStr != "" {
				privatePaths = append(privatePaths, strings.Split(privatePathStr, ";")...)
			}

			hasProbing := false

			libsDir = strings.TrimPrefix(libsDir, "./")
			libsDir = strings.TrimPrefix(libsDir, ".\\")

			for _, privatePath := range privatePaths {
				if privatePath == libsDir {
					hasProbing = true
					break
				}
			}

			if !hasProbing {
				privatePaths = append(privatePaths, libsDir)
				probing.CreateAttr("privatePath", strings.Join(privatePaths, ";"))
			}
		}

		doc.WriteSettings.UseCRLF = true
		bytes, _ := doc.WriteToBytes()

		if err := ioutil.WriteFile(exeConfig, bytes, 0666); err != nil {
			log.LogError(fmt.Errorf("fix exe.config failed: %s : %s", exeConfig, err.Error()), false)
		}
	}

	dir := filepath.Dir(exeConfig)

	// additional dlls
	if files, err := util.ReadAllFile(dir); err == nil {
		for _, file := range files {
			if !strings.HasSuffix(file, ".dll") {
				continue
			}

			exists := false

			for _, deps := range allDeps {
				if file == deps.Name {
					exists = true
					break
				}
			}

			if exists {
				continue
			}

			allDeps = append(allDeps, Deps{
				Name:       file,
				Path:       file,
				SecondPath: file,
				Type:       Assembly,
				Locale:     "",
			})
		}
	}

	// additional satellite assemblies
	if sdir, err := util.ReadAllDir(dir); err == nil {
		for _, d := range sdir {
			if files, err := util.ReadAllFile(filepath.Join(dir, d)); err == nil {
				for _, file := range files {
					if strings.HasSuffix(file, ".resources.dll") {
						allDeps = append(allDeps, Deps{
							Name:       file,
							Path:       d + "/" + file,
							SecondPath: d + "/" + file,
							Type:       Resource,
							Locale:     d,
						})
					}
				}
			}
		}
	}

	return allDeps, true
}

// FixRuntimeConfig 添加libs到runtimeconfig.json
func FixRuntimeConfig(runtimeConfig string, libsDir string, subDirs []string, srmMapping map[string]string, sharedRuntimeMode bool, usePatch bool, useWPF bool) bool {
	jsonBytes, err := ioutil.ReadFile(runtimeConfig)
	if err != nil {
		log.LogError(fmt.Errorf("can not read runtimeconfig.json: %s : %s", runtimeConfig, err.Error()), false)
		return false
	}

	json, err := simplejson.NewJson(jsonBytes)
	if err != nil {
		log.LogError(fmt.Errorf("invalid runtimeconfig.json: %s : %s", runtimeConfig, err.Error()), false)
		return false
	}

	libsDir = strings.ReplaceAll(libsDir, "\\", "/")

	libsDir = strings.TrimSuffix(libsDir, "/")

	libsDirs := make([]string, 0)

	libsDirs = append(libsDirs, ".")
	libsDirs = append(libsDirs, libsDir)

	for _, v := range subDirs {
		libsDirs = append(libsDirs, libsDir+"/"+strings.ReplaceAll(v, "\\", "/"))
	}

	json.SetPath([]string{
		"runtimeOptions",
		"configProperties",
		"NetBeautyLibsDir",
	}, strings.Join(libsDirs, ";"))

	var appID = ""

	if sharedRuntimeMode {
		parts := strings.Split(strings.ReplaceAll(runtimeConfig, "\\", "/"), "/")
		fileName := parts[len(parts)-1]
		entry := strings.Split(fileName, ".runtimeconfig.")[0]
		appID, _ = util.GetStringMD5(entry)

		json.SetPath([]string{
			"runtimeOptions",
			"configProperties",
			"NetBeautyAppID",
		}, appID)

		srmMappingArr := make([]string, 0)
		for fileName, md5 := range srmMapping {
			srmMappingArr = append(srmMappingArr, fileName+":"+md5)
		}
		srmMappingStr := strings.Join(srmMappingArr, "|")
		json.SetPath([]string{
			"runtimeOptions",
			"configProperties",
			"NetBeautySharedRuntimeMode",
		}, "default")
		json.SetPath([]string{
			"runtimeOptions",
			"configProperties",
			"NetBeautySharedRuntimeMapping",
		}, srmMappingStr)
	} else {
		json.SetPath([]string{
			"runtimeOptions",
			"configProperties",
			"NetBeautySharedRuntimeMode",
		}, "no")
	}

	if usePatch {
		runtimeOptions := json.Get("runtimeOptions")
		additionalProbingPaths, ok := runtimeOptions.CheckGet("additionalProbingPaths")

		var existPaths []string = []string{}

		if ok {
			existPaths, err = additionalProbingPaths.StringArray()
			if err != nil {
				log.LogPanic(fmt.Errorf("invalid runtimeconfig.json: %s", err.Error()), 1)
			}
		}

		var pathsHashTable map[string]bool = make(map[string]bool, len(existPaths))

		for _, path := range existPaths {
			pathsHashTable[path] = true
		}

		var addPaths []string = []string{}

		if !sharedRuntimeMode {
			addPaths = append(addPaths, libsDir)
		}

		srmNativeDir := libsDir + "/srm_native/" + appID

		if sharedRuntimeMode {
			for fileName, md5 := range srmMapping {
				if strings.Contains(fileName, "/") {
					// // resources
					// addPaths = append(addPaths, strings.Join([]string{
					// 	libsDir,
					// 	"locales",
					// 	fileName,
					// 	md5,
					// }, "/"))
				} else {
					if fileName == "System.Collections.dll" ||
						fileName == "System.Memory.dll" ||
						fileName == "System.Private.CoreLib.dll" ||
						fileName == "System.Runtime.dll" ||
						fileName == "System.Runtime.Extensions.dll" ||
						fileName == "System.Runtime.InteropServices.dll" ||
						fileName == "System.Runtime.InteropServices.RuntimeInformation.dll" ||
						fileName == "System.Runtime.Loader.dll" ||
						fileName == "System.IO.FileSystem.dll" ||
						fileName == "System.IO.Packaging.dll" {
						addPaths = append(addPaths, strings.Join([]string{
							libsDir,
							fileName,
							md5,
						}, "/"))
					}

					if useWPF {
						if fileName == "PresentationCore.dll" ||
							fileName == "PresentationFramework.dll" ||
							fileName == "WindowsBase.dll" ||
							fileName == "System.Xaml.dll" {
							addPaths = append(addPaths, strings.Join([]string{
								libsDir,
								fileName,
								md5,
							}, "/"))
						}
					}
				}
			}
		}

		for _, path := range addPaths {
			pathsHashTable[path] = true
		}

		var resultPaths []string = []string{}

		for path := range pathsHashTable {
			if path == "" {
				continue
			}

			resultPaths = append(resultPaths, path)
		}

		// NOTE: SRM模式下，dll存在二级结构，libsDir必须置于最后去搜索
		// 否则将会直接将libsDir下dll二级目录当成已存在dll去读取
		if sharedRuntimeMode {
			resultPaths = append([]string{srmNativeDir}, resultPaths...)
			resultPaths = append(resultPaths, libsDir)
		}

		runtimeOptions.Set("additionalProbingPaths", resultPaths)
	}

	jsonBytes, _ = json.EncodePretty()
	if err := ioutil.WriteFile(runtimeConfig, jsonBytes, 0666); err != nil {
		log.LogError(fmt.Errorf("add NetBeautyLibsDir to runtimeconfig.json failed: %s : %s", runtimeConfig, err.Error()), false)
		return false
	}

	return true
}

// FindFXRVersion 从deps.json中提取出FXR Version
func FindFXRVersion(deps string) (string, string) {
	fxrVersion, rid := "", ""

	jsonBytes, err := ioutil.ReadFile(deps)
	if err != nil {
		return "", ""
	}

	json, err := simplejson.NewJson(jsonBytes)
	if err != nil {
		return "", ""
	}

	// targets
	targets, _ := json.Get("targets").Map()
	for _, target := range targets {
		for targetName := range target.(map[string]interface{}) {
			// 解析出fxr信息
			if !strings.HasPrefix(targetName, "runtime") {
				continue
			}
			isResolver := strings.Contains(targetName, "Microsoft.NETCore.DotNetHostResolver")
			isRuntime := strings.Contains(targetName, "Microsoft.NETCore.App.Runtime")

			if !isResolver && !isRuntime {
				continue
			}

			patterns := []string{
				// 2.x
				`^runtime.([\w\-\.]+).Microsoft.NETCore.DotNetHostResolver/([\w\-\.]+)$`,
				// 3.0.x
				`^runtimepack.runtime.([\w\-\.]+).Microsoft.NETCore.DotNetHostResolver/([\w\-\.]+)$`,
				// ≥3.1.x
				`^runtimepack.Microsoft.NETCore.App.Runtime.([\w\-\.]+)/([\w\-\.]+)$`,
			}

			for _, pattern := range patterns {
				regex, _ := regexp.Compile(pattern)
				matches := regex.FindStringSubmatch(targetName)

				if len(matches) == 3 {
					rid = matches[1]
					fxrVersion = matches[2]

					return "v" + fxrVersion, rid
				}
			}
		}
	}

	return "", ""
}

// FixDeps 分析deps.json中的依赖项
func FixDeps(deps string, entry string, enableDebug bool, usePatch bool, sharedRuntimeMode bool) ([]Deps, bool, bool) {
	var isAspNetCore = false
	var useWPF = false
	var verifyWpfDllSet = false

	var webConfig = "web.config"
	var windowsBaseDll = "WindowsBase.dll"
	var presentationCoreDll = "PresentationCore.dll"

	var allAnalyzedDeps = make([]analyzedDeps, 0)
	var allDeps = make([]Deps, 0)

	dir := filepath.Dir(deps)

	jsonBytes, err := ioutil.ReadFile(deps)
	if err != nil {
		log.LogError(fmt.Errorf("can not read deps.json: %s : %s", deps, err.Error()), false)
		return allDeps, useWPF, isAspNetCore
	}

	json, err := simplejson.NewJson(jsonBytes)
	if err != nil {
		log.LogError(fmt.Errorf("invalid deps.json: %s : %s", deps, err.Error()), false)
		return allDeps, useWPF, isAspNetCore
	}

	var shouldSkip = func(fileName string, entry string) bool {
		// entry point
		if fileName == entry+".dll" ||
			strings.Contains(fileName, "hostfxr.") ||
			strings.Contains(fileName, "hostpolicy.") {
			return true
		}

		// clr
		if isAspNetCore || !usePatch {
			if strings.Contains(fileName, "clrjit.") ||
				strings.Contains(fileName, "coreclr.") {
				return true
			}
		}

		// nbloader
		if !usePatch {
			if fileName == "nbloader.dll" {
				return true
			}
		}

		// nbloader dependencies
		if !usePatch {
			if fileName == "System.Collections.dll" ||
				fileName == "System.Memory.dll" ||
				fileName == "System.Private.CoreLib.dll" ||
				fileName == "System.Runtime.dll" ||
				fileName == "System.Runtime.Extensions.dll" ||
				fileName == "System.Runtime.InteropServices.dll" ||
				fileName == "System.Runtime.InteropServices.RuntimeInformation.dll" ||
				fileName == "System.Runtime.Loader.dll" ||
				fileName == "System.IO.FileSystem.dll" ||
				fileName == "System.IO.Packaging.dll" ||
				strings.Contains(fileName, "libSystem.Native") {
				return true
			}
		}

		// ASP.NET Core
		if isAspNetCore {
			if strings.Contains(fileName, "aspnetcore") ||
				strings.Contains(fileName, "aspnetcorev2") {
				return true
			}
		}

		// WPF
		if !usePatch && useWPF {
			if fileName == "PresentationFramework.dll" ||
				fileName == "WindowsBase.dll" ||
				fileName == "System.Xaml.dll" {
				return true
			}
		}

		// additional WPF
		if !usePatch && verifyWpfDllSet {
			if fileName == "PresentationCore.dll" ||
				strings.Contains(fileName, "PresentationNative_") ||
				strings.Contains(fileName, "wpfgfx_") ||
				strings.Contains(fileName, "vcruntime") ||
				strings.Contains(fileName, "D3DCompiler_") ||
				strings.Contains(fileName, "PenImc_") ||
				strings.Contains(fileName, "PenImc2_") {
				return true
			}
		}

		return false
	}

	targets, _ := json.Get("targets").Map()
	for _, target := range targets {
		for depsName, depsObj := range target.(map[string]interface{}) {
			if depsName == "nbloader" {
				continue
			}

			runtime := depsObj.(map[string]interface{})["runtime"]
			if runtime != nil {
				for filePath := range runtime.(map[string]interface{}) {
					filePath2 := strings.ReplaceAll(filePath, "\\", "/")
					parts := strings.Split(filePath2, "/")
					fileName := parts[len(parts)-1]

					if fileName == presentationCoreDll {
						useWPF = true
					}

					allAnalyzedDeps = append(allAnalyzedDeps, analyzedDeps{
						Category:   runtime.(map[string]interface{}),
						ItemKey:    filePath,
						Name:       fileName,
						Path:       fileName,
						SecondPath: fileName,
						Type:       Assembly,
						Locale:     "",
					})
				}
			}

			resources := depsObj.(map[string]interface{})["resources"]
			if resources != nil {
				for filePath, locale := range resources.(map[string]interface{}) {
					filePath2 := strings.ReplaceAll(filePath, "\\", "/")
					parts := strings.Split(filePath2, "/")
					fileName := parts[len(parts)-1]
					culture := locale.(map[string]interface{})["locale"].(string)

					allAnalyzedDeps = append(allAnalyzedDeps, analyzedDeps{
						Category:   resources.(map[string]interface{}),
						ItemKey:    filePath,
						Name:       fileName,
						Path:       culture + "/" + fileName,
						SecondPath: culture + "/" + fileName,
						Type:       Resource,
						Locale:     culture,
					})
				}
			}

			native := depsObj.(map[string]interface{})["native"]
			if native != nil {
				for filePath := range native.(map[string]interface{}) {
					filePath2 := strings.ReplaceAll(filePath, "\\", "/")
					parts := strings.Split(filePath2, "/")
					fileName := parts[len(parts)-1]

					allAnalyzedDeps = append(allAnalyzedDeps, analyzedDeps{
						Category:   native.(map[string]interface{}),
						ItemKey:    filePath,
						Name:       fileName,
						Path:       fileName,
						SecondPath: filePath2,
						Type:       Native,
						Locale:     "",
					})
				}
			}
		}
	}

	webConfigPath := dir + "/" + webConfig

	if util.PathExists(webConfigPath) {
		isAspNetCore = true
	}

	if isAspNetCore {
		log.LogDetail("ASP.NET Core: Yes")
	} else {
		log.LogDetail("ASP.NET Core: No")
	}

	windowsBaseDllPath := dir + "/" + windowsBaseDll

	if useWPF && util.PathExists(windowsBaseDllPath) {
		content, err := ioutil.ReadFile(windowsBaseDllPath)
		if err != nil {
			log.LogError(fmt.Errorf("read dll failed: %s : %s", windowsBaseDllPath, err.Error()), true)
		}
		verifyWpfDllSet = bytes.Index(content, []byte("VerifyWpfDllSet")) != -1
	}

	if useWPF {
		log.LogDetail("Use WPF: Yes")

		if verifyWpfDllSet {
			log.LogDetail("VerifyWpfDllSet: Yes")
		} else {
			log.LogDetail("VerifyWpfDllSet: No")
		}
	} else {
		log.LogDetail("Use WPF: No")
	}

	if enableDebug {
		log.LogDetail("Enable Debugging: Yes")
	} else {
		log.LogDetail("Enable Debugging: No")
	}

	for _, analyzed := range allAnalyzedDeps {
		if shouldSkip(analyzed.Name, entry) {
			continue
		}

		allDeps = append(allDeps, Deps{
			Name:       analyzed.Name,
			Path:       analyzed.Path,
			SecondPath: analyzed.SecondPath,
			Type:       analyzed.Type,
			Locale:     analyzed.Locale,
		})

		// debug files
		if !enableDebug {
			if strings.Contains(analyzed.Name, "mscordaccore") ||
				strings.Contains(analyzed.Name, "mscordbi") {
				if !strings.HasPrefix(analyzed.ItemKey, "./") {
					delete(analyzed.Category, analyzed.ItemKey)
				}
				continue
			}
		}

		if usePatch {
			var needRooted = true
			if sharedRuntimeMode {
				needRooted = false

				if analyzed.Type == Native {
					needRooted = true
				}

				if analyzed.Name == "System.Collections.dll" ||
					analyzed.Name == "System.Memory.dll" ||
					analyzed.Name == "System.Private.CoreLib.dll" ||
					analyzed.Name == "System.Runtime.dll" ||
					analyzed.Name == "System.Runtime.Extensions.dll" ||
					analyzed.Name == "System.Runtime.InteropServices.dll" ||
					analyzed.Name == "System.Runtime.InteropServices.RuntimeInformation.dll" ||
					analyzed.Name == "System.Runtime.Loader.dll" ||
					analyzed.Name == "System.IO.FileSystem.dll" ||
					analyzed.Name == "System.IO.Packaging.dll" {
					needRooted = true
				}

				if analyzed.Name == "PresentationCore.dll" || analyzed.Name == "PresentationFramework.dll" ||
					analyzed.Name == "WindowsBase.dll" ||
					analyzed.Name == "System.Xaml.dll" {
					needRooted = true
				}
			}

			if needRooted {
				if analyzed.Type == Resource {
					analyzed.Category["./"+analyzed.Locale+"/"+analyzed.Name] = map[string]interface{}{
						"locale": analyzed.Locale,
					}
				} else {
					analyzed.Category["./"+analyzed.Name] = make(map[string]interface{})
				}
			}
		}

		if !strings.HasPrefix(analyzed.ItemKey, "./") {
			delete(analyzed.Category, analyzed.ItemKey)
		}
	}

	if usePatch {
		libraries, _ := json.Get("libraries").Map()
		for k, lib := range libraries {
			fixLib := lib.(map[string]interface{})
			fixLib["path"] = "./"
			libraries[k] = fixLib
		}
		json.Set("libraries", libraries)
	}

	jsonBytes, _ = json.EncodePretty()
	if err := ioutil.WriteFile(deps, jsonBytes, 0666); err != nil {
		log.LogError(fmt.Errorf("fix deps.json failed: %s : %s", deps, err.Error()), false)
	}

	// additional satellite assemblies
	if sdir, err := util.ReadAllDir(dir); err == nil {
		for _, d := range sdir {
			if files, err := util.ReadAllFile(filepath.Join(dir, d)); err == nil {
				for _, file := range files {
					if strings.HasSuffix(file, ".resources.dll") {
						allDeps = append(allDeps, Deps{
							Name:       file,
							Path:       d + "/" + file,
							SecondPath: d + "/" + file,
							Type:       Resource,
							Locale:     d,
						})
					}
				}
			}
		}
	}

	return allDeps, useWPF, isAspNetCore
}

func onlinePath() string {
	return GitCDN + "/raw/" + GitTree
}

func artifactsOnlinePath() string {
	return onlinePath() + "/artifacts"
}

func artifactsVersionOldURL() string {
	return artifactsOnlinePath() + artifactsVersionTXT
}

func artifactsVersionURL() string {
	return artifactsOnlinePath() + artifactsVersionJSON
}

func runtimeJSONPath(specific string) string {
	return path.Join(localArtifactsPath, specific)
}

func runtimeCompatibilityJSONPath() string {
	return runtimeJSONPath(runtimeCompatibilityJSONName)
}

func runtimeSupportedJSONPath() string {
	return runtimeJSONPath(runtimeSupportedJSONName)
}

func runtimeJSONURL(specific string) string {
	return artifactsOnlinePath() + "/" + specific
}

func runtimeCompatibilityJSONURL() string {
	return runtimeJSONURL(runtimeCompatibilityJSONName)
}

func runtimeSupportedJSONURL() string {
	return runtimeJSONURL(runtimeSupportedJSONName)
}

func artifactFile(version string, rid string) string {
	return path.Join(localArtifactsPath, version, rid+".Release", GetHostFXRNameByRID(rid))
}

// GetHostFXRNameByRID 根据RID取hostfxr文件名
func GetHostFXRNameByRID(rid string) string {
	if strings.Contains(rid, "win") {
		return "hostfxr.dll"
	} else if strings.Contains(rid, "osx") {
		return "libhostfxr.dylib"
	}
	return "libhostfxr.so"
}

func readJSON(path string, errlog bool) *simplejson.Json {
	bytes, err := ioutil.ReadFile(path)
	if err != nil && errlog {
		log.LogInfo(fmt.Sprintf("read json failed: %s : %s", path, err.Error()))
		return nil
	}
	json, err := simplejson.NewJson(bytes)
	if err != nil && errlog {
		log.LogDetail(fmt.Sprintf("parse json failed: %s : %s", path, err.Error()))
		return nil
	}
	return json
}

func readLocalArtifactsVersionJSON() map[string]interface{} {
	json := readJSON(artifactsVersionPath, false)
	if json == nil {
		return nil
	}
	localVersions, err := json.Map()
	if err == nil {
		return localVersions
	}
	errMsg := formatError(getLocalArtifactsVersionErr, errors.New("invalid artifactsVersion Json: "+artifactsVersionPath))
	log.LogPanic(errors.New(errMsg), 1)
	return nil
}

func updateLocalArtifactsVersionJSON(data map[string]interface{}) bool {
	if !util.EnsureDirExists(localArtifactsPath, 0777) {
		log.LogError(fmt.Errorf(pathNotWriteableErr, localArtifactsPath), false)
		return false
	}

	json := simplejson.New()
	for k, v := range data {
		json.Set(k, v)
	}

	jsonBytes, err := json.EncodePretty()
	if err != nil {
		log.LogError(fmt.Errorf(encodeJSONErr, err.Error()), false)
		return false
	}
	err = ioutil.WriteFile(artifactsVersionPath, jsonBytes, 0666)
	if err != nil {
		log.LogError(fmt.Errorf(pathNotWriteableErr, artifactsVersionPath), false)
	}
	return err == nil
}

func verid(version string, rid string) string {
	return version + "/" + rid
}

// GetLocalArtifactsVersion 获取本地补丁版本
func GetLocalArtifactsVersion(version string, rid string) string {
	localVersions := readLocalArtifactsVersionJSON()
	if localVersions != nil {
		for verid, localVer := range localVersions {
			// verid: version/rid
			localVerStr := localVer.(string)
			s := strings.Split(verid, "/")
			if version == s[0] && rid == s[1] {
				return localVerStr
			}
		}
	}
	return ""
}

// GetOnlineArtifactsVersion 获取线上补丁版本
func GetOnlineArtifactsVersion(version string, rid string) string {
	// 如果缓存存在则尝试读取，如果缓存找不到就直接返回（缓存必然是最新的）
	var readCache = func() string {
		if onlineVersionCache != nil {
			json, success := onlineVersionCache.CheckGet(verid(version, rid))
			if success {
				return json.MustString("")
			}
		}
		return ""
	}
	if onlineVersionCache != nil {
		return readCache()
	}

	var latest = false

	http.DefaultClient.Timeout = 5 * time.Second
	if response, err := http.Get(artifactsVersionOldURL()); err == nil && response.StatusCode == 200 {
		defer response.Body.Close()
		if bytes, err := ioutil.ReadAll(response.Body); err == nil {
			onlineVersion := string(bytes)
			// 读入本地版本号
			oldVersion := ""
			if oldVerBytes, err := ioutil.ReadFile(artifactsVersionOldPath); err == nil {
				oldVersion = string(oldVerBytes)
			}

			// 判断版本号
			latest = oldVersion == onlineVersion

			if !latest {
				// 写入本地版本号
				if err := ioutil.WriteFile(artifactsVersionOldPath, bytes, 0666); err != nil {
					log.LogError(err, false)
				}
			}
		}
	}

	// 加载本地缓存版本库
	if latest && util.PathExists(onlineArtifactsVersionPath) {
		onlineVersionCache = readJSON(onlineArtifactsVersionPath, true)
		return readCache()
	}

	// 如果本地不是最新的就获取网上最新的版本号
	// 获取版本超时短一点可减少网络环境差所造成的影响
	http.DefaultClient.Timeout = 10 * time.Second
	if response, err := http.Get(artifactsVersionURL()); err == nil && response.StatusCode == 200 {
		defer response.Body.Close()
		if bytes, err := ioutil.ReadAll(response.Body); err == nil {
			onlineVersionCache, _ = simplejson.NewJson(bytes)
			// 写入本地缓存
			if err := ioutil.WriteFile(onlineArtifactsVersionPath, bytes, 0666); err != nil {
				log.LogError(err, false)
			}
			return readCache()
		}
	}

	return readCache()
}

func getLocalRuntimeCompatibilityVersion() string {
	return GetLocalArtifactsVersion("runtime", "compatibility")
}

func getLocalRuntimeSupportedVersion() string {
	return GetLocalArtifactsVersion("runtime", "supported")
}

func getOnlineRuntimeCompatibilityVersion() string {
	return GetOnlineArtifactsVersion("runtime", "compatibility")
}

func getOnlineRuntimeSupportedVersion() string {
	return GetOnlineArtifactsVersion("runtime", "supported")
}

// CheckRunConfigJSON 检查本地runtimeConfig，自动下载最新（强制性）
func CheckRunConfigJSON() {
	log.LogInfo("checking runtime.*.json version...")
	onlineCVersion := getOnlineRuntimeCompatibilityVersion()
	onlineSVersion := getOnlineRuntimeSupportedVersion()
	if onlineCVersion == "" {
		log.LogDetail("fetch online runtime compatibility version failed")
		return
	}
	if onlineSVersion == "" {
		log.LogDetail("fetch online runtime supported version failed")
		return
	}
	localCVersion := getLocalRuntimeCompatibilityVersion()
	localSVersion := getLocalRuntimeSupportedVersion()
	var mapping = map[string][2]string{
		runtimeCompatibilityJSONName: {localCVersion, onlineCVersion},
		runtimeSupportedJSONName:     {localSVersion, onlineSVersion},
	}
	for name, vers := range mapping {
		if vers[0] == vers[1] {
			log.LogInfo(fmt.Sprintf("%s no need to update", name))
			continue
		}
		log.LogDetail(fmt.Sprintf("updating %s...", name))
		url := runtimeJSONURL(name)
		path := runtimeJSONPath(name)
		specific := strings.TrimSuffix(strings.TrimPrefix(name, "runtime."), ".json")
		if !DownloadFile(url, path) || !WriteLocalArtifactsVersion("runtime", specific, vers[1]) {
			log.LogDetail(fmt.Sprintf("update %s failed", name))
		} else {
			log.LogInfo(fmt.Sprintf("update %s succeeded", name))
		}
	}
}

// FindCompatibleRID 匹配线上所支持的RID
func FindCompatibleRID(rid string) string {
	runtimeCompatibilityJSON := readJSON(runtimeCompatibilityJSONPath(), true)
	if runtimeCompatibilityJSON == nil {
		return ""
	}
	crids, _ := runtimeCompatibilityJSON.Get(rid).StringArray()
	if crids == nil || len(crids) == 0 {
		return ""
	}
	return crids[0]
}

// DownloadFile 下载文件
func DownloadFile(url string, des string) bool {
	http.DefaultClient.Timeout = timeout

	response, err := http.Get(url)
	if err == nil && response.StatusCode == 200 {
		defer response.Body.Close()
		if bytes, err := ioutil.ReadAll(response.Body); err != nil {
			log.LogError(err, false)
		} else {
			des = strings.ReplaceAll(des, "\\", "/")
			path := path.Dir(des)
			if !util.EnsureDirExists(path, 0777) {
				log.LogError(fmt.Errorf(pathNotWriteableErr, path), false)
			} else {
				f, err := os.Create(des)
				defer f.Close()
				log.LogError(err, false)
				if err == nil {
					if _, err := f.Write(bytes); err == nil {
						return true
					}
					log.LogError(err, false)
				}
			}
		}
	}
	return false
}

// DownloadArtifact 下载指定版本、RID的补丁
func DownloadArtifact(version string, rid string) bool {
	fileName := GetHostFXRNameByRID(rid)
	artifactURL := fmt.Sprintf("%s/%s/%s.Release/%s", artifactsOnlinePath(), version, rid, fileName)

	artifactFile := path.Join(localArtifactsPath, version, rid+".Release", fileName)

	return DownloadFile(artifactURL, artifactFile)
}

// WriteLocalArtifactsVersion 更新本地补丁版本
func WriteLocalArtifactsVersion(fxrVersion string, rid string, version string) bool {
	if !util.EnsureDirExists(localArtifactsPath, 0777) {
		log.LogError(fmt.Errorf(pathNotWriteableErr, localArtifactsPath), false)
		return false
	}
	var json map[string]interface{}
	if util.PathExists(artifactsVersionPath) {
		json = readLocalArtifactsVersionJSON()
	} else {
		json = make(map[string]interface{})
	}
	key := verid(fxrVersion, rid)
	if version == "" {
		delete(json, key)
	} else {
		json[key] = version
	}
	return updateLocalArtifactsVersionJSON(json)
}

// CopyArtifactTo 复制补丁到指定文件夹
func CopyArtifactTo(version string, rid string, des string) bool {
	if !IsLocalArtifactExists(version, rid) {
		log.LogError(fmt.Errorf("Artifact does not exist. %s/%s", version, rid), false)
		return false
	}
	artifactName := GetHostFXRNameByRID(rid)
	artifactFile := artifactFile(version, rid)
	des = path.Join(path.Clean(des), artifactName)
	if _, err := util.CopyFile(artifactFile, des); err != nil {
		log.LogError(fmt.Errorf("Cannot copy artifact from %s to %s. %s", artifactFile, des, err.Error()), false)
	}
	return true
}

// IsLocalArtifactExists 判断本地是否存在某个版本的补丁
func IsLocalArtifactExists(version string, rid string) bool {
	return util.PathExists(artifactFile(version, rid))
}

// SetCDN 设置默认CDN
func SetCDN(cdn string) bool {
	if err := ioutil.WriteFile(gitCDNPath, []byte(cdn), 0666); err != nil {
		log.LogError(err, false)
		return false
	}
	return true
}

// GetCDN 获取默认CDN
func GetCDN() string {
	if gitcdn, err := ioutil.ReadFile(gitCDNPath); err == nil {
		return string(gitcdn)
	}
	return ""
}

// DelCDN 删除默认CDN
func DelCDN() bool {
	if err := os.Remove(gitCDNPath); err != nil {
		log.LogError(err, false)
		return false
	}
	return true
}
