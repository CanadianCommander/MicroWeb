package main

import (
	"errors"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/CanadianCommander/MicroWeb/pkg/cache"
	"github.com/CanadianCommander/MicroWeb/pkg/logger"
)

const FILE_READ_BUFFER_SIZE = 0xFFFF //64 KB
const TEMPLATE_FILE_EXTENSION = ".gohtml"

func HandleResourceRequest(res http.ResponseWriter, req *http.Request) {
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
		} else {
			//read and serve file
			buff := ReadFileToBuff(fsPath)
			if buff != nil {
				res.Write((*buff)[:])
			} else {
				res.WriteHeader(500)
				return false
			}
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
		} else {
			// real file path
			buff := ReadFileToBuff(fsPath)
			if buff != nil {
				return plugin.HandleRequest(req, res, buff)
			} else {
				res.WriteHeader(500)
				return false
			}
		}
	}

	return true
}

func ReadFileToBuff(fsPath string) *[]byte {
	cacheBuffer := cache.FetchFromCache(cache.CACHE_RESOURCE, fsPath)
	if cacheBuffer != nil {
		//cache hit
		logger.LogVerbose("Loading resource from cache: %s", fsPath)
		return cacheBuffer.(*[]byte)
	} else {
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
			var buff []byte = make([]byte, FILE_READ_BUFFER_SIZE)
			bytesOut, _ = file.Read(buff)
			copy(byteBuffer[byteBufferIndex:], buff[:bytesOut])
			byteBufferIndex += bytesOut
		}

		cache.AddToCache(cache.CACHE_RESOURCE, fsPath, (&byteBuffer))
		return &byteBuffer
	}
}

/**
  takes a url and resolves it to a file system path, if possible.
*/
func URLToFilesystem(url string) (string, error) {
	if strings.Compare(url, "/") == 0 {
		return "", errors.New("Invalid Path")
	}

	templatePath := path.Join(globalSettings.GetStaticResourcePath(), url)

	// if some how url contains '..' characters we could accidentally expose the entire filesystem
	// make sure we are still within the static resource path
	if !strings.Contains(templatePath, path.Clean(globalSettings.GetStaticResourcePath())) {
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
		templatePath, err = URLToFilesystem(path.Join(url, "index.gohtml"))
		if err != nil {
			//attempt to redirect to standard html index file
			templatePath, err = URLToFilesystem(path.Join(url, "index.html"))
			if err != nil {
				logger.LogInfo("Requsted resource is directory and nether \"index.gohtml\" nor \"index.html\" exist in it")
				return templatePath, errors.New("File not Found")
			}
		}
	}

	return templatePath, nil
}
