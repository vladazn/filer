package files

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"sync"
)

type Settings struct {
	FileNames    []string
	DownloadPath string
}

var lockers map[string]*sync.RWMutex
var downloadPath string

func Setup(params Settings) {
	lockers = make(map[string]*sync.RWMutex)
	if params.DownloadPath == "" {
		downloadPath, _ = os.Getwd()
	} else {
		downloadPath = params.DownloadPath
	}

	if len(params.FileNames) > 0 {
		createLock(params.FileNames)
	}
}

func createLock(names []string) {
	for _, name := range names {
		lockers[name] = &sync.RWMutex{}
	}
}

func Write(data interface{}, fileName string, format string, dir string) error {
	if fileName == "" {
		return errors.New("file name should be specified")
	}
	lockers[fileName].Lock()
	defer lockers[fileName].Unlock()

	if dir == "" {
		dir = downloadPath
	}
	dir = dir + "/" + fileName + "." + format
	dataToWrite, _ := json.Marshal(data)
	err := ioutil.WriteFile(dir, dataToWrite, 0644)
	if err != nil {
		return err
	}
	return nil
}

func Read(readChan chan []byte, stopChan chan bool,  fileName string, format string, dir string) error {
	if fileName == "" {
		return errors.New("file name should be specified")
	}
	lockers[fileName].RLock()
	defer lockers[fileName].RUnlock()


	if dir == "" {
		dir = downloadPath
	}
	dir = dir + "/" + fileName + "." + format

	file, err := os.Open(dir)
	if err != nil {
		return err
	}

	for {
		chunk := make([]byte, 1024)
		bytesRead, err := file.Read(chunk)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		readChan <- chunk[:bytesRead]
		//
		//data := &pb.Chunk{
		//	Content: chunk[:bytesRead],
		//}
		//(*stream).Send(data)
	}
	stopChan<-true
	return nil
}
