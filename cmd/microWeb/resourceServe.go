package main

import (
	"errors"
	"html/template"
	"microWeb/pkg/logger"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

const FILE_READ_BUFFER_SIZE = 0xFFFF //64 KB
const TEMPLATE_FILE_EXTENSION = ".gohtml"

func HandleTemplateRequest(res http.ResponseWriter, req *http.Request) {
	serveTime := time.Now()
	logger.LogVerbose("%s request from %s for URL %s", req.Method, req.RemoteAddr, req.URL)

	if handleTemplateRequest(res, req) {
		logger.LogVerbose("Served Request: %s, to %s in %f ms", req.URL.Path, req.RemoteAddr, float64(time.Since(serveTime).Nanoseconds())/1000000.0)
	} else {
		logger.LogWarning("failed to serve request, [%s] to %s", req.URL.Path, req.RemoteAddr)
	}
}

func handleTemplateRequest(res http.ResponseWriter, req *http.Request) bool {

	fsPath, fsErr := URLToFilesystem(req.URL.Path)
	if fsErr != nil {
		logger.LogInfo("Resource not found: %s", req.URL.Path)
		res.WriteHeader(404)
		return false
	}

	if path.Ext(fsPath) == TEMPLATE_FILE_EXTENSION {
		//read, parse and serve template file
		pluginToUse, pErr := GetPluginByResourcePath(fsPath)
		if pErr != nil {
			logger.LogError("Template file has no plugin associated with it! file: %s", fsPath)
			res.WriteHeader(500)
			return false
		}
		plugin, pErr := LoadTemplatePlugin(pluginToUse)
		if pErr != nil {
			logger.LogError("Template plugin failed to load")
			res.WriteHeader(500)
			return false
		}

		buff := ReadFileToBuff(fsPath)
		if buff != nil {
			templateParser := template.New("root")
			_, tErr := templateParser.Parse(string((*buff)[:]))
			if tErr != nil {
				logger.LogError("Template Parsing error: %s. template file: %s", tErr.Error(), fsPath)
				res.WriteHeader(500)
				return false
			}

			tData, pErr := plugin.GetTemplateStruct(req)
			if pErr != nil {
				logger.LogError("Plugin Triggered abort: %s", pErr.Error())
				res.WriteHeader(500)
				return false
			}

			templateParser.Execute(res, tData)
		} else {
			res.WriteHeader(500)
			return false
		}

	} else {
		//read and serve normal file
		buff := ReadFileToBuff(fsPath)
		if buff != nil {
			res.Write((*buff)[:])
		} else {
			res.WriteHeader(500)
			return false
		}
	}

	return true
}

func ReadFileToBuff(fsPath string) *[]byte {
	cacheBuffer := FetchFromCache(CACHE_RESOURCE, fsPath)
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

		AddToCache(CACHE_RESOURCE, fsPath, (&byteBuffer))
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
