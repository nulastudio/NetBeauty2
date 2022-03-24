package manager

import (
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
	files, err := filepath.Glob(path.Join(dir, "*runtimeconfig*.json"))
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
func FixRuntimeConfig(runtimeConfig string, libsDir string, subDirs []string) bool {
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

	jsonBytes, _ = json.EncodePretty()
	if err := ioutil.WriteFile(runtimeConfig, jsonBytes, 0666); err != nil {
		log.LogError(fmt.Errorf("add NetBeautyLibsDir to runtimeconfig.json failed: %s : %s", runtimeConfig, err.Error()), false)
		return false
	}

	return true
}

// FixDeps 分析deps.json中的依赖项
func FixDeps(deps string, entry string) []Deps {
	var allDeps = make([]Deps, 0)
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
			fileName == "PresentationFramework.dll" ||
			fileName == "WindowsBase.dll" ||
			fileName == "System.Xaml.dll" ||
			fileName == "System.Private.CoreLib.dll" ||
			fileName == "System.Runtime.dll" ||
			fileName == "System.Runtime.Extensions.dll" ||
			fileName == "System.Runtime.InteropServices.dll" ||
			fileName == "System.Runtime.Loader.dll" ||
			fileName == "System.IO.FileSystem.dll" ||
			strings.Contains(fileName, "libSystem.Native") ||
			strings.Contains(fileName, "clrjit.") ||
			strings.Contains(fileName, "coreclr.") ||
			strings.Contains(fileName, "mscordaccore") || // for adebugging
			strings.Contains(fileName, "mscordbi") || // for adebugging
			strings.Contains(fileName, "hostfxr.") ||
			strings.Contains(fileName, "hostpolicy.") {
			return true
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
					filePath = strings.ReplaceAll(filePath, "\\", "/")
					parts := strings.Split(filePath, "/")
					fileName := parts[len(parts)-1]

					if shouldSkip(fileName, entry) {
						continue
					}

					allDeps = append(allDeps, Deps{
						Name:       fileName,
						Path:       fileName,
						SecondPath: fileName,
						Type:       Assembly,
					})

					delete(runtime.(map[string]interface{}), filePath)
				}
			}

			resources := depsObj.(map[string]interface{})["resources"]
			if resources != nil {
				for filePath, locale := range resources.(map[string]interface{}) {
					filePath = strings.ReplaceAll(filePath, "\\", "/")
					parts := strings.Split(filePath, "/")
					fileName := parts[len(parts)-1]
					culture := locale.(map[string]interface{})["locale"].(string)

					allDeps = append(allDeps, Deps{
						Name:       fileName,
						Path:       culture + "/" + fileName,
						SecondPath: culture + "/" + fileName,
						Type:       Resource,
					})

					delete(resources.(map[string]interface{}), filePath)
				}
			}

			native := depsObj.(map[string]interface{})["native"]
			if native != nil {
				for filePath := range native.(map[string]interface{}) {
					filePath = strings.ReplaceAll(filePath, "\\", "/")
					parts := strings.Split(filePath, "/")
					fileName := parts[len(parts)-1]

					if shouldSkip(fileName, entry) {
						continue
					}

					allDeps = append(allDeps, Deps{
						Name:       fileName,
						Path:       fileName,
						SecondPath: filePath,
						Type:       Native,
					})

					delete(native.(map[string]interface{}), filePath)
				}
			}
		}
	}

	jsonBytes, _ = json.EncodePretty()
	if err := ioutil.WriteFile(deps, jsonBytes, 0666); err != nil {
		log.LogError(fmt.Errorf("fix deps.json failed: %s : %s", deps, err.Error()), false)
	}

	// additional satellite assemblies
	dir := filepath.Dir(deps)
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
