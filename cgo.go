package main

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// import "C"

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

//export initialize
func initialize(videoId string) *CgoDownloadState {
	di := NewDownloadInfo()
	di.VideoID = videoId
	di.URL = fmt.Sprintf("https://www.youtube.com/watch?v=%s", di.VideoID)
	return &CgoDownloadState{
		info: di,
	}
}

//export registerFormat
func registerFormat(p *CgoDownloadState, id, fmtUrl, manifestUrl, filepath string) {
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
func loadCookies(cookieFile string) bool {
	cjar, err := info.ParseNetscapeCookiesFile(cookieFile)
	if err != nil {
		LogError("Failed to load cookies file: %s", err)
		return false
	}

	client.Jar = cjar
	LogInfo("Loaded cookie file %s", cookieFile)
	return true
}

//export runDownloader
func runDownloader(state *CgoDownloadState) {
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
func interrupt(state *CgoDownloadState) {
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
func poll(state *CgoDownloadState, timeout int) string {
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
}
