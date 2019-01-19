package main

import (
	"errors"
	"mime"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/CanadianCommander/MicroWeb/pkg/cache"
	"github.com/CanadianCommander/MicroWeb/pkg/logger"
	mwsettings "github.com/CanadianCommander/MicroWeb/pkg/mwSettings"
)

const fileReadBufferSize = 0xFFFF //64 KB
const templateFileExt = ".gohtml"

/*
HandleRequest is called to handle any and all http requests made by clients
*/
func HandleRequest(res http.ResponseWriter, req *http.Request) {
	serveTime := time.Now()
	logger.LogVerbose("%s request from %s for URL %s", req.Method, req.RemoteAddr, req.URL)

	if handleRequest(res, req) {
		logger.LogVerbose("Served Request: %s, to %s in %f ms", req.URL.Path, req.RemoteAddr, float64(time.Since(serveTime).Nanoseconds())/1000000.0)
	} else {
		logger.LogWarning("failed to serve request, [%s] to %s", req.URL.Path, req.RemoteAddr)
	}
}

func handleRequest(res http.ResponseWriter, req *http.Request) bool {

	fsPath, fsErr := URLToFilesystem(req.URL.Path)
	pluginToUse, pErr := GetPluginByResourcePath(fsPath)
	if pErr != nil {
		//no plugin just serve resource raw
		if fsErr != nil {
			//file not found
			logger.LogInfo("Resource not found: %s", req.URL.Path)
			res.WriteHeader(404)
			return false
		}
		//read and serve file
		buff := ReadFileToBuff(fsPath)
		if buff != nil {
			mimeType := mime.TypeByExtension(path.Ext(fsPath))
			res.Header().Add("Content-Type", mimeType)
			res.Header().Add("Cache-Control", "max-age="+mwsettings.GetSettingString("tune/max-age"))
			res.Write((*buff)[:])
		} else {
			res.WriteHeader(500)
			return false
		}

	} else {
		//push resorce file through plugin
		plugin, pErr := LoadPlugin(pluginToUse)
		if pErr != nil {
			logger.LogError("Plugin failed to load")
			res.WriteHeader(500)
			return false
		}

		if fsErr != nil {
			// virtual file path
			return plugin.HandleVirtualRequest(req, res)
		}

		return plugin.HandleRequest(req, res, fsPath)
	}

	return true
}

/*
ReadFileToBuff reads the enter file found at fsPath in to a []byte buffer and returns it.
If any thing goes wrong nil is returned.

Note. This function uses the global cache provided by the,
"github.com/CanadianCommander/MicroWeb/pkg/cache" package.
*/
func ReadFileToBuff(fsPath string) *[]byte {
	cacheBuffer := cache.FetchFromCache(cache.CacheTypeResource, fsPath)
	if cacheBuffer != nil {
		//cache hit
		logger.LogVerbose("Loading resource from cache: %s", fsPath)
		return cacheBuffer.(*[]byte)
	}
	//cache miss
	logger.LogVerbose("Loading resource from file: %s", fsPath)
	file, err := os.Open(fsPath)
	defer file.Close()
	if err != nil {
		logger.LogError("Could not open resource file at: %s ", fsPath)
		return nil
	}
	fileInfo, _ := file.Stat()
	byteBuffer := make([]byte, fileInfo.Size())

	bytesOut := 1
	byteBufferIndex := 0
	for bytesOut != 0 {
		var buff = make([]byte, fileReadBufferSize)
		bytesOut, _ = file.Read(buff)
		copy(byteBuffer[byteBufferIndex:], buff[:bytesOut])
		byteBufferIndex += bytesOut
	}

	cache.AddToCache(cache.CacheTypeResource, fsPath, (&byteBuffer))
	return &byteBuffer
}

/*
URLToFilesystem takes a url and resolves it to a file system path, if possible,
taking in to account the global setting for static resource path.
*/
func URLToFilesystem(url string) (string, error) {
	webRoot := mwsettings.GetSettingString("general/staticDirectory")
	templatePath := path.Join(webRoot, url)

	// if some how url contains '..' characters we could accidentally expose the entire filesystem
	// make sure we are still within the static resource path
	if !strings.Contains(templatePath, path.Clean(mwsettings.GetSettingString("general/staticDirectory"))) {
		logger.LogWarning("Suspicius URL activity. URL resolved to: %s", templatePath)
		return "", errors.New("URL invalid")
	}

	fInfo, err := os.Stat(templatePath)
	if err != nil {
		logger.LogInfo("Requested resource: %s Not found", templatePath)
		return templatePath, err
	}
	if fInfo.IsDir() {
		logger.LogVerbose("Requested resource is directory. Redirecting to index file")
		//attempt to redirect to go index file
		templatePath = path.Join(webRoot, path.Join(url, "index.gohtml"))
		fInfo, err = os.Stat(templatePath)
		if err != nil {
			//attempt to redirect to standard html index file
			templatePath = path.Join(webRoot, path.Join(url, "index.html"))
			fInfo, err = os.Stat(templatePath)
			if err != nil {
				logger.LogInfo("Requsted resource is directory and nether \"index.gohtml\" nor \"index.html\" exist in it")
				return templatePath, errors.New("File not Found")
			}
		}
	}

	return templatePath, nil
}
