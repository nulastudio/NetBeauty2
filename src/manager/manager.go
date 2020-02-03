package manager

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
	"time"

	"github.com/bitly/go-simplejson"

	log "github.com/nulastudio/NetCoreBeauty/src/log"
	util "github.com/nulastudio/NetCoreBeauty/src/util"
)

// GitCDN git仓库镜像（默认为github）
var GitCDN = "https://github.com/nulastudio/HostFXRPatcher"

// Logger 日志记录器
var Logger = log.DefaultLogger

var timeout = 60 * time.Second

var localPath = filepath.Clean(os.TempDir()) + "/NetCoreBeauty"
var localArtifactsPath = localPath + "/artifacts"
var artifactsVersionTXT = "/ArtifactsVersion.txt"
var artifactsVersionJSON = "/ArtifactsVersion.json"
var artifactsVersionPath = localArtifactsPath + artifactsVersionJSON

var runtimeCompatibilityJSONName = "runtime.compatibility.json"
var runtimeSupportedJSONName = "runtime.supported.json"

var pathNotWriteableErr = "cannot create path or path is not writeable: %s"
var getLocalArtifactsVersionErr = "get local artifacts version failed: %s"
var encodeJSONErr = "cannot encode json: %s"

var onlineVersionCache string = ""

func formatError(format string, err error) string {
	return fmt.Sprintf(format, err)
}

func onlinePath() string {
	return GitCDN + "/raw/master"
}

func artifactsOnlinePath() string {
	return onlinePath() + "/artifacts"
}

func artifactsVersionURL() string {
	return artifactsOnlinePath() + artifactsVersionTXT
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
	bytes, err := ioutil.ReadFile(artifactsVersionPath)
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

func getLocalRuntimeCompatibilityVersion() string {
	return GetLocalArtifactsVersion("runtime", "compatibility")
}

func getLocalRuntimeSupportedVersion() string {
	return GetLocalArtifactsVersion("runtime", "supported")
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
	key := fxrVersion + "/" + rid
	if version == "" {
		delete(json, key)
	} else {
		json[key] = version
	}
	return updateLocalArtifactsVersionJSON(json)
}

// GetOnlineArtifactsVersion 获取线上补丁版本
func GetOnlineArtifactsVersion() string {
	if onlineVersionCache != "" {
		return onlineVersionCache
	}
	// 获取版本超时短一点可减少网络环境差所造成的影响
	http.DefaultClient.Timeout = 10 * time.Second
	if response, err := http.Get(artifactsVersionURL()); err == nil && response.StatusCode == 200 {
		defer response.Body.Close()
		if bytes, err := ioutil.ReadAll(response.Body); err == nil {
			onlineVersionCache = string(bytes)
			return onlineVersionCache
		}
	}
	// 获取失败就清除缓存吧
	onlineVersionCache = ""
	return onlineVersionCache
}

// IsLocalArtifactExists 判断本地是否存在某个版本的补丁
func IsLocalArtifactExists(version string, rid string) bool {
	return util.PathExists(artifactFile(version, rid))
}

// CheckRunConfigJSON 检查本地runtimeConfig，自动下载最新（强制性）
func CheckRunConfigJSON() {
	log.LogInfo("checking runtime.*.json version...")
	onlineVersion := GetOnlineArtifactsVersion()
	if onlineVersion == "" {
		log.LogDetail("checking runtime.*.json version failed")
		return
	}
	localCVersion := getLocalRuntimeCompatibilityVersion()
	localSVersion := getLocalRuntimeSupportedVersion()
	var mapping = map[string]string{
		runtimeCompatibilityJSONName: localCVersion,
		runtimeSupportedJSONName:     localSVersion,
	}
	for name, version := range mapping {
		if version != onlineVersion {
			log.LogDetail(fmt.Sprintf("updating %s...", name))
			url := runtimeJSONURL(name)
			path := runtimeJSONPath(name)
			specific := strings.TrimSuffix(strings.TrimPrefix(name, "runtime."), ".json")
			if !DownloadFile(url, path) || !WriteLocalArtifactsVersion("runtime", specific, onlineVersion) {
				log.LogDetail(fmt.Sprintf("update %s failed", name))
			} else {
				log.LogInfo(fmt.Sprintf("update %s succeeded", name))
			}
		} else {
			log.LogInfo(fmt.Sprintf("%s no need to update", name))
		}
	}
}

// FindCompatibleRID 匹配线上所支持的RID
func FindCompatibleRID(rid string) string {
	runtimeCompatibilityJSON := readJSON(runtimeCompatibilityJSONPath(), true)
	runtimeSupportedJSON := readJSON(runtimeSupportedJSONPath(), true)
	if runtimeCompatibilityJSON == nil || runtimeSupportedJSON == nil {
		return rid
	}
	crids, _ := runtimeCompatibilityJSON.Get(rid).StringArray()
	srids, _ := runtimeSupportedJSON.StringArray()
	for _, crid := range crids {
		for _, srid := range srids {
			if crid == srid {
				return srid
			}
		}
	}
	return rid
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

// DeleteArtifact 下载指定版本、RID的补丁
func DeleteArtifact(version string, rid string) bool {
	artifactFile := artifactFile(version, rid)
	ret := WriteLocalArtifactsVersion(version, rid, "")
	if !ret {
		return false
	}
	if util.PathExists(artifactFile) {
		return os.Remove(artifactFile) == nil
	}
	return true
}

// UpdateArtifact 更新指定版本、RID的补丁
func UpdateArtifact(version string, rid string) bool {
	onlineVersion := GetOnlineArtifactsVersion()
	// 为了避免残留过时缓存，强制删除再下载
	return DeleteArtifact(version, rid) &&
		DownloadArtifact(version, rid) &&
		WriteLocalArtifactsVersion(version, rid, onlineVersion)
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

// FixRuntimeConfig 添加additionalProbingPaths
func FixRuntimeConfig(runtimeConfigFile string, libsDir string) bool {
	jsonBytes, err := ioutil.ReadFile(runtimeConfigFile)
	if err != nil {
		log.LogError(fmt.Errorf("can not read runtimeconfig.json: %s", err.Error()), false)
		return false
	}
	json, err := simplejson.NewJson(jsonBytes)
	if err != nil {
		log.LogPanic(fmt.Errorf("invalid runtimeconfig.json: %s", err.Error()), 1)
	}

	var found = false
	runtimeOptions, ok := json.CheckGet("runtimeOptions")
	if !ok {
		runtimeOptions = simplejson.New()
		json.Set("runtimeOptions", runtimeOptions)
	}
	additionalProbingPaths, ok := runtimeOptions.CheckGet("additionalProbingPaths")
	var paths []string = []string{}
	if ok {
		paths, err = additionalProbingPaths.StringArray()
		if err != nil {
			log.LogPanic(fmt.Errorf("invalid runtimeconfig.json: %s", err.Error()), 1)
		}
	}
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
	jsonBytes, err = json.EncodePretty()
	if err != nil {
		log.LogPanic(fmt.Errorf("can not encode runtimeconfig.json: %s", err.Error()), 1)
	}
	err = ioutil.WriteFile(runtimeConfigFile, jsonBytes, 0666)
	if err != nil {
		log.LogError(fmt.Errorf("can not write runtimeconfig.json: %s", err.Error()), false)
	}
	return err == nil
}

// FixDeps 修改deps.json
func FixDeps(deps string) ([]string, string, string) {
	jsonBytes, err := ioutil.ReadFile(deps)
	if err != nil {
		log.LogError(fmt.Errorf("can not open deps.json: %s : %s", deps, err.Error()), false)
		return nil, "", ""
	}

	json, err := simplejson.NewJson(jsonBytes)
	if err != nil {
		log.LogError(fmt.Errorf("invalid deps.json: %s : %s", deps, err.Error()), false)
		return nil, "", ""
	}

	files := []string{}
	rid := ""
	fxrVersion := ""

	// targets
	targets, _ := json.Get("targets").Map()
	for _, target := range targets {
		for targetName, depsObj := range target.(map[string]interface{}) {
			// 解析出fxr信息
			if strings.HasPrefix(targetName, "runtime") &&
				(strings.Contains(targetName, "Microsoft.NETCore.DotNetHostResolver") ||
					strings.Contains(targetName, "Microsoft.NETCore.App.Runtime")) {
				// 2.x
				regex, _ := regexp.Compile("^runtime\\.([\\w\\-]+)\\.Microsoft\\.NETCore\\.DotNetHostResolver\\/([\\w\\-\\.]+)$")
				matches := regex.FindStringSubmatch(targetName)
				if len(matches) == 0 {
					// 3.0.x
					regex, _ = regexp.Compile("^runtimepack\\.runtime\\.([\\w\\-]+)\\.Microsoft\\.NETCore\\.DotNetHostResolver\\/([\\w\\-\\.]+)$")
					matches = regex.FindStringSubmatch(targetName)
					if len(matches) == 0 {
						// 3.1.x
						regex, _ = regexp.Compile("^runtimepack\\.Microsoft\\.NETCore\\.App\\.Runtime\\.([\\w\\-]+)\\/([\\w\\-\\.]+)$")
						matches = regex.FindStringSubmatch(targetName)
					}
				}
				if len(matches) == 3 {
					rid = matches[1]
					fxrVersion = matches[2]
				}
				log.LogInfo(fmt.Sprintf("fxr v%s/%s detected in %s", fxrVersion, rid, deps))
			}
			if depsObj != nil {
				components := map[string]int{
					// NOTE: runtimeTargets未确认是否需要处理
					// "runtimeTargets": 1,
					"runtime":   1,
					"native":    1,
					"compile":   1,
					"resources": 2,
				}
				for cname, segments := range components {
					component := depsObj.(map[string]interface{})[cname]
					if component != nil {
						newComponent := make(map[string]interface{})
						for k := range component.(map[string]interface{}) {
							components := strings.Split(strings.ReplaceAll(k, "\\", "/"), "/")
							length := len(components)
							fileName := strings.Join(components[length-segments:], "/")
							files = append(files, fileName)
							newComponent["./"+fileName] = make(map[string]interface{})
						}
						depsObj.(map[string]interface{})[cname] = newComponent
					}
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
	if err := ioutil.WriteFile(deps, jsonBytes, 0666); err != nil {
		log.LogError(fmt.Errorf("fix deps.json failed: %s : %s", deps, err.Error()), false)
		return nil, "", ""
	}

	if fxrVersion == "" || rid == "" {
		log.LogError(fmt.Errorf("incomplete fxr info [%s/%s] found in deps.json: %s", fxrVersion, rid, deps), false)
	}

	return files, "v" + fxrVersion, rid
}
