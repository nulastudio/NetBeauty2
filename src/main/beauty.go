package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bitly/go-simplejson"
)

var workingDir, _ = os.Getwd()
var beautyDir = ""
var libsDir = "libs"
var disablePatch = false
var gitcdn = "https://github.com/nulastudio/HostFXRPatcher"

func main() {
	// arguments check
	if len(os.Args) < 2 {
		fmt.Println("no beauty path specify")
		fmt.Println("")
		help()
		os.Exit(0)
	} else {
		beautyDir = os.Args[1]
		if len(os.Args) >= 3 {
			libsDir = os.Args[2]
		}
		beautyDir = strings.Trim(beautyDir, "\"")
		libsDir = strings.Trim(libsDir, "\"")
		beautyDir, _ = filepath.Abs(beautyDir)
		if len(os.Args) >= 4 {
			disablePatch = os.Args[3] == "True"
		}
		if len(os.Args) >= 5 {
			gitcdn = os.Args[4]
		}
	}

	fmt.Println("running ncbeauty...")

	// fix runtimeconfig.json
	fmt.Println("fixing runtimeconfig.json...")
	runtimeConfigs, _ := filepath.Glob(beautyDir + "/*runtimeconfig*.json")
	for _, runtimeConfig := range runtimeConfigs {
		fixRuntimeConfig(runtimeConfig, libsDir)
	}
	fmt.Println("runtimeconfig.json fixed")

	// fix deps.json
	fmt.Println("fixing deps.json...")
	dependencies, _ := filepath.Glob(beautyDir + "/*deps.json")
	for _, deps := range dependencies {
		deps = strings.ReplaceAll(deps, "\\", "/")
		mainProgram := strings.Replace(path.Base(deps), "deps.json", "", -1)
		depsFiles := fixDependencies(deps, mainProgram)
		// 移动文件
		for _, depsFile := range depsFiles {
			if strings.Contains(depsFile, mainProgram) {
				continue
			}
			if strings.Contains(depsFile, "hostfxr") {
				continue
			}
			if strings.Contains(depsFile, "hostpolicy") {
				continue
			}

			absDepsFile := path.Join(beautyDir, depsFile)
			absDesFile := path.Join(beautyDir, libsDir, depsFile)
			oldPath := path.Dir(absDepsFile)
			newPath := path.Dir(absDesFile)
			if pathExists(absDepsFile) {
				if !pathExists(newPath) {
					os.MkdirAll(newPath, 0777)
				}
				os.Rename(absDepsFile, absDesFile)

				fileName := strings.TrimSuffix(path.Base(depsFile), path.Ext(depsFile))
				extFiles := []string{".pdb", ".xml"}
				for _, extFile := range extFiles {
					oldFile := path.Join(oldPath, fileName+extFile)
					newFile := path.Join(newPath, fileName+extFile)
					if pathExists(oldFile) {
						os.Rename(oldFile, newFile)
					}
				}
				dir, _ := ioutil.ReadDir(oldPath)
				if len(dir) == 0 {
					os.Remove(oldPath)
				}
			}
		}
	}
	fmt.Println("deps.json fixed")
	fmt.Println("ncbeauty done. Enjoy it!")
}

func help() {
	fmt.Println("Usage:")
	fmt.Println("ncbeauty <beautyDir> [<LibsDir>] [<disablePatch=True|False>] [<gitcdn>]")
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func fixRuntimeConfig(runtimeConfig string, libsDir string) {
	jsonBytes, _ := ioutil.ReadFile(runtimeConfig)
	json, _ := simplejson.NewJson(jsonBytes)

	var found = false
	runtimeOptions := json.Get("runtimeOptions")
	paths, _ := runtimeOptions.Get("additionalProbingPaths").StringArray()
	for _, path := range paths {
		if path == libsDir {
			found = true
			break
		}
	}
	if !found {
		paths = append(paths, libsDir)
	}
	runtimeOptions.Set("additionalProbingPaths", paths)
	jsonBytes, _ = json.EncodePretty()
	ioutil.WriteFile(runtimeConfig, jsonBytes, 0666)
}

func fixDependencies(deps string, mainProgram string) []string {
	jsonBytes, _ := ioutil.ReadFile(deps)
	json, _ := simplejson.NewJson(jsonBytes)
	files := []string{}
	rid := ""
	fxrVersion := ""
	fxrName := ""

	// targets
	targets, _ := json.Get("targets").Map()
	for _, target := range targets {
		for targetName, depsObj := range target.(map[string]interface{}) {
			// 解析出fxr信息
			if strings.HasPrefix(targetName, "runtime") && strings.Contains(targetName, "Microsoft.NETCore.DotNetHostResolver") {
				regex, _ := regexp.Compile("^runtime\\.([\\w-]+)\\.Microsoft\\.NETCore\\.DotNetHostResolver\\/([\\d\\.]+)$")
				matches := regex.FindStringSubmatch(targetName)
				rid = matches[1]
				fxrVersion = matches[2]
				fmt.Printf("fxr %s.%s detected\n", rid, fxrVersion)
			}
			if depsObj != nil {
				// runtime
				runtimes := depsObj.(map[string]interface{})["runtime"]
				if runtimes != nil {
					newRuntimes := make(map[string]interface{})
					for k := range runtimes.(map[string]interface{}) {
						components := strings.Split(strings.ReplaceAll(k, "\\", "/"), "/")
						fileName := components[len(components)-1]
						files = append(files, fileName)
						newRuntimes["./"+fileName] = make(map[string]interface{})
					}
					depsObj.(map[string]interface{})["runtime"] = newRuntimes
				}
				// NOTE: runtimeTargets未确认是否需要处理
				// runtimeTargets
				// runtimeTargets := depsObj.(map[string]interface{})["runtimeTargets"]
				// if runtimeTargets != nil {
				// 	newRuntimeTargets := make(map[string]interface{})
				// 	for k := range runtimeTargets.(map[string]interface{}) {
				// 		components := strings.Split(strings.ReplaceAll(k, "\\", "/"), "/")
				// 		fileName := components[len(components)-1]
				// 		files = append(files, fileName)
				// 		newRuntimeTargets["./"+fileName] = make(map[string]interface{})
				// 	}
				// 	depsObj.(map[string]interface{})["runtimeTargets"] = newRuntimeTargets
				// }
				// native
				natives := depsObj.(map[string]interface{})["native"]
				if natives != nil {
					newNatives := make(map[string]interface{})
					for k := range natives.(map[string]interface{}) {
						components := strings.Split(strings.ReplaceAll(k, "\\", "/"), "/")
						fileName := components[len(components)-1]
						files = append(files, fileName)
						newNatives["./"+fileName] = make(map[string]interface{})
						if strings.Contains(fileName, "hostfxr") {
							fxrName = fileName
						}
					}
					depsObj.(map[string]interface{})["native"] = newNatives
				}
				// compile
				compiles := depsObj.(map[string]interface{})["compile"]
				if compiles != nil {
					newCompiles := make(map[string]interface{})
					for k := range compiles.(map[string]interface{}) {
						components := strings.Split(strings.ReplaceAll(k, "\\", "/"), "/")
						fileName := components[len(components)-1]
						files = append(files, fileName)
						newCompiles["./"+fileName] = make(map[string]interface{})
					}
					depsObj.(map[string]interface{})["compile"] = newCompiles
				}
				// resources
				resources := depsObj.(map[string]interface{})["resources"]
				if resources != nil {
					newResources := make(map[string]interface{})
					for k := range resources.(map[string]interface{}) {
						components := strings.Split(strings.ReplaceAll(k, "\\", "/"), "/")
						fileName := components[len(components)-2] + "/" + components[len(components)-1]
						files = append(files, fileName)
						newResources["./"+fileName] = make(map[string]interface{})
					}
					depsObj.(map[string]interface{})["resources"] = newResources
				}
			}
		}
	}
	json.Set("targets", targets)

	// libraries
	libraries, _ := json.Get("libraries").Map()
	for k, lib := range libraries {
		fixLib := lib.(map[string]interface{})
		fixLib["path"] = "./"
		libraries[k] = fixLib
	}
	json.Set("libraries", libraries)

	jsonBytes, _ = json.EncodePretty()
	ioutil.WriteFile(deps, jsonBytes, 0666)

	// patch
	if disablePatch {
		fmt.Println("hostfxr patch has been disable, skipped.")
		return files
	}
	fmt.Println("patching hostfxr...")
	// TODO: match the right rid
	fmt.Printf("downloading patched hostfxr: %s.%s\n", rid, fxrVersion)
	fxrURL := fmt.Sprintf("%s/raw/master/artifacts/v%s/%s.Release/%s", gitcdn, fxrVersion, rid, fxrName)
	response, err := http.Get(fxrURL)
	var fxrData []byte
	if err == nil {
		if response.StatusCode != 200 {
			err = errors.New(response.Status)
		} else {
			fxrData, err = ioutil.ReadAll(response.Body)
		}
	}
	if err != nil {
		fmt.Println("download patch failed due to:")
		// ensure download
		fmt.Println(err)
		os.Exit(1)
	}
	absFxrName := path.Join(beautyDir, fxrName)
	absFxrBakName := absFxrName + ".bak"
	fmt.Printf("backuping fxr to %s\n", absFxrBakName)
	if pathExists(absFxrBakName) {
		fmt.Println("fxr backup found, skipped.")
	} else {
		err := os.Rename(absFxrName, absFxrBakName)
		if err != nil {
			fmt.Println("backup failed.")
			fmt.Println(err)
			return files
		}
	}
	f, err := os.Create(absFxrName)
	if err != nil {
		fmt.Printf("open %s failed\n", absFxrName)
		fmt.Println(err)
		return files
	}
	bytes, err := f.Write(fxrData)
	if err != nil || bytes == 0 {
		fmt.Printf("write %s failed\n", absFxrName)
		fmt.Println(err)
		return files
	}
	fmt.Println("patch succeed")

	return files
}
