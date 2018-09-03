package pluginUtil

import (
	"html/template"
	"io"
	"os"

	"github.com/CanadianCommander/MicroWeb/pkg/cache"
	"github.com/CanadianCommander/MicroWeb/pkg/logger"
)

const FILE_READ_BUFFER_SIZE = 0xFFFF //64 KB

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

func ProcessTemplate(templateFileBuffer *[]byte, out io.Writer, tStruct interface{}) error {
	templateParser := template.New("root")
	_, tErr := templateParser.Parse(string((*templateFileBuffer)[:]))
	if tErr != nil {
		logger.LogError("could not parse template file w/ error: %s", tErr.Error())
		return tErr
	}

	return templateParser.Execute(out, tStruct)
}
