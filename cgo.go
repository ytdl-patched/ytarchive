//go:build cgo

package main

// #include <stdio.h>
// #include <stdlib.h>

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
	"unsafe"
)

import "C"

// unused variables needed to build
const (
	ActionAsk = iota
	ActionDo
	ActionDoNot
)

var (
	downloadThumbnail bool
	mkv               bool
	addMeta           bool
	info              *DownloadInfo
)

// to prevent from the instance to be collected by GC
var instances map[uintptr]*CgoDownloadState

func stateToPtr(st *CgoDownloadState) uintptr {
	p := uintptr(unsafe.Pointer(st))

    if instances == nil {
        instances = make(map[uintptr]*CgoDownloadState)
    }
    instances[p] = st

    return p
}

func ptrToState(ptr uintptr) *CgoDownloadState {
	return (*CgoDownloadState)(unsafe.Pointer(ptr))
}

type CgoDownloadState struct {
	info         *DownloadInfo
	formats      []YtdlMicroformat
	progressChan chan *ProgressInfo
	// this chan will receive dataTypes (effectively formatId) instead
	dlDoneChan chan string
}

// subset of yt-dl(p)'s format dict, speficially segmented formats
type YtdlMicroformat struct {
	id          string
	url         string
	manifestUrl string
	filepath    string
}

//export goInfo
func goInfo() *C.char {
	var a uintptr
	// order: <size of uintptr>
	return C.CString(fmt.Sprintf("%d", unsafe.Sizeof(a)))
}

//export initialize
func initialize(videoIdC *C.char) uintptr {
	videoId := C.GoString(videoIdC)
	di := NewDownloadInfo()
	di.VideoID = videoId
	di.URL = fmt.Sprintf("https://www.youtube.com/watch?v=%s", di.VideoID)
	return stateToPtr(&CgoDownloadState{
		info: di,
	})
}

//export release
func release(ptr uintptr) {
	if _, ok := instances[ptr]; ok {
		delete(instances, ptr)
	}
}

//export registerFormat
func registerFormat(ptr uintptr, idC, fmtUrlC, manifestUrlC, filepathC *C.char) {
	id := C.GoString(idC)
	fmtUrl := C.GoString(fmtUrlC)
	manifestUrl := C.GoString(manifestUrlC)
	filepath := C.GoString(filepathC)
	p := ptrToState(ptr)
	p.formats = append(p.formats, YtdlMicroformat{
		id:          id,
		url:         fmtUrl,
		manifestUrl: manifestUrl,
		filepath:    filepath,
	})

	p.info.MDLInfo[id] = &MediaDLInfo{
		ActiveJobs: 0,
		// actually it's not used for download, we'll always download at filepath
		BasePath: "",
		DataType: id,
	}
	p.info.SetDownloadUrl(id, fmtUrl)

	p.info.Jobs += 1
}

//export loadCookies
func loadCookies(cookieFileC *C.char) C.int {
	cookieFile := C.GoString(cookieFileC)
	cjar, err := info.ParseNetscapeCookiesFile(cookieFile)
	if err != nil {
		LogError("Failed to load cookies file: %s", err)
		return C.int(0) // false
	}

	client.Jar = cjar
	LogInfo("Loaded cookie file %s", cookieFile)
	return C.int(1) // true
}

//export runDownloader
func runDownloader(ptr uintptr) {
	state := ptrToState(ptr)
	state.progressChan = make(chan *ProgressInfo, state.info.Jobs*2)
	state.dlDoneChan = make(chan string, state.info.Jobs)
	dlDoneChan := make(chan struct{}, state.info.Jobs)

	var wg sync.WaitGroup
	for _, fmt := range state.formats {
		LogInfo("Starting download for %s to %s", fmt.id, fmt.filepath)
		wg.Add(1)
		go func(id, filepass string, pc chan *ProgressInfo, ddc chan string) {
			defer wg.Done()
			defer func() { ddc <- id }()
			state.info.DownloadStream(id, filepass, pc, dlDoneChan)
		}(fmt.id, fmt.filepath, state.progressChan, state.dlDoneChan)
	}
	wg.Wait()
}

//export interrupt
func interrupt(ptr uintptr) {
	state := ptrToState(ptr)
	state.info.Stop()
	// we flag all other formats too
	for _, v := range state.formats {
		state.info.SetFinished(v.id)
	}
}

func serializeError(err error) string {
	data := map[string]interface{}{
		"type":   "error",
		"params": fmt.Sprintf("%s", err),
	}
	res, _ := json.Marshal(data)
	return string(res)
}

//export poll
func poll(ptr uintptr, timeoutC C.int) *C.char {
	state := ptrToState(ptr)
	timeout := int(timeoutC)

	ret := func() string {
		select {
		case p := <-state.progressChan:
			data := map[string]interface{}{
				"type":   "progress",
				"params": *p,
			}
			res, err := json.Marshal(data)
			if err != nil {
				return serializeError(err)
			}
			return string(res)
		case fmtId := <-state.dlDoneChan:
			data := map[string]interface{}{
				"type":   "done",
				"params": fmtId,
			}
			res, err := json.Marshal(data)
			if err != nil {
				return serializeError(err)
			}
			return string(res)
		case <-time.After(time.Duration(timeout) * time.Millisecond):
			data := map[string]interface{}{
				"type":   "tryagain",
				"params": "",
			}
			res, err := json.Marshal(data)
			if err != nil {
				return serializeError(err)
			}
			return string(res)
		}
	}()

	return C.CString(ret)
}

func main() {}
