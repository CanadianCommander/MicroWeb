package pluginUtil

import (
	"html/template"
	"io"
	"os"

	"github.com/CanadianCommander/MicroWeb/pkg/cache"
	"github.com/CanadianCommander/MicroWeb/pkg/logger"
)

const fileReadBufferSize = 0xFFFF //64 KB

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
ProcessTemplate takes the template described by templateFileBuffer and uses the html/template
package to parse and execute the template, pushing output on the, out io.Writer.
*/
func ProcessTemplate(templateFileBuffer *[]byte, out io.Writer, tStruct interface{}) error {
	templateParser := template.New("root")
	_, tErr := templateParser.Parse(string((*templateFileBuffer)[:]))
	if tErr != nil {
		logger.LogError("could not parse template file w/ error: %s", tErr.Error())
		return tErr
	}

	return templateParser.Execute(out, tStruct)
}
