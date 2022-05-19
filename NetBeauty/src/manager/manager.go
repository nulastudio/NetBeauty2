package manager

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"

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
}

type Deps struct {
	Name       string
	Path       string
	SecondPath string
	Type       DepsType
}

// Logger 日志记录器
var Logger = log.DefaultLogger

func formatError(format string, err error) string {
	return fmt.Sprintf(format, err)
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

// FindRuntimeConfigJSON 寻找指定目录下的*runtimeconfig*.json
func FindRuntimeConfigJSON(dir string) []string {
	files, err := filepath.Glob(filepath.Join(dir, "*runtimeconfig*.json"))
	if err != nil {
		log.LogDetail(formatError("find runtimeconfig.json failed: %s", err))
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

// FixRuntimeConfig 添加libs到runtimeconfig.json
func FixRuntimeConfig(runtimeConfig string, libsDir string, subDirs []string, srmMapping map[string]string, sharedRuntimeMode bool) bool {
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

	if sharedRuntimeMode {
		parts := strings.Split(strings.ReplaceAll(runtimeConfig, "\\", "/"), "/")
		fileName := parts[len(parts)-1]
		entry := strings.Split(fileName, ".runtimeconfig.")[0]
		appID, _ := util.GetStringMD5(entry)

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

	jsonBytes, _ = json.EncodePretty()
	if err := ioutil.WriteFile(runtimeConfig, jsonBytes, 0666); err != nil {
		log.LogError(fmt.Errorf("add NetBeautyLibsDir to runtimeconfig.json failed: %s : %s", runtimeConfig, err.Error()), false)
		return false
	}

	return true
}

// FixDeps 分析deps.json中的依赖项
func FixDeps(deps string, entry string, enableDebug bool) []Deps {
	var useWPF = false
	var verifyWpfDllSet = false

	var windowsBaseDll = "WindowsBase.dll"

	var allAnalyzedDeps = make([]analyzedDeps, 0)
	var allDeps = make([]Deps, 0)

	dir := filepath.Dir(deps)

	jsonBytes, err := ioutil.ReadFile(deps)
	if err != nil {
		log.LogError(fmt.Errorf("can not read deps.json: %s : %s", deps, err.Error()), false)
		return allDeps
	}

	json, err := simplejson.NewJson(jsonBytes)
	if err != nil {
		log.LogError(fmt.Errorf("invalid deps.json: %s : %s", deps, err.Error()), false)
		return allDeps
	}

	var shouldSkip = func(fileName string, entry string) bool {
		if fileName == entry+".dll" ||
			fileName == "nbloader.dll" ||
			fileName == "PresentationFramework.dll" || // for GUI
			fileName == "WindowsBase.dll" || // for GUI
			fileName == "System.Xaml.dll" || // for GUI
			fileName == "System.Collections.dll" || // for nbloader
			fileName == "System.Memory.dll" || // for nbloader
			fileName == "System.Private.CoreLib.dll" || // for nbloader
			fileName == "System.Runtime.dll" || // for nbloader
			fileName == "System.Runtime.Extensions.dll" || // for nbloader
			fileName == "System.Runtime.InteropServices.dll" || // for nbloader
			fileName == "System.Runtime.Loader.dll" || // for nbloader
			fileName == "System.IO.FileSystem.dll" || // for nbloader
			fileName == "System.IO.Packaging.dll" || // for nbloader
			strings.Contains(fileName, "libSystem.Native") || // for nbloader
			strings.Contains(fileName, "aspnetcore") || // for ASP.NET Core
			strings.Contains(fileName, "aspnetcorev2") || // for ASP.NET Core
			(enableDebug && strings.Contains(fileName, "mscordaccore")) || // for debugging
			(enableDebug && strings.Contains(fileName, "mscordbi")) || // for debugging
			strings.Contains(fileName, "clrjit.") ||
			strings.Contains(fileName, "coreclr.") ||
			strings.Contains(fileName, "hostfxr.") ||
			strings.Contains(fileName, "hostpolicy.") {
			return true
		}

		if verifyWpfDllSet {
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

					if fileName == windowsBaseDll {
						useWPF = true
					}

					allAnalyzedDeps = append(allAnalyzedDeps, analyzedDeps{
						Category:   runtime.(map[string]interface{}),
						ItemKey:    filePath,
						Name:       fileName,
						Path:       fileName,
						SecondPath: fileName,
						Type:       Assembly,
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
					})
				}
			}
		}
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
	} else {
		log.LogDetail("Use WPF: No")
	}

	if verifyWpfDllSet {
		log.LogDetail("VerifyWpfDllSet: Yes")
	} else {
		log.LogDetail("VerifyWpfDllSet: No")
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
		})

		delete(analyzed.Category, analyzed.ItemKey)
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
						})
					}
				}
			}
		}
	}

	return allDeps
}
