package hls

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"path"
	"strings"
)

type Downloader struct {
	masterManifestSrc string
	baseManifestUrl   *url.URL
	renditions        []Rendition
	client            *http.Client
}

func NewDownloader(masterManifestSrc string) Downloader {
	uri, err := url.Parse(masterManifestSrc)
	if err != nil {
		panic("invalid master manifest src format")
	}

	uri.Path = path.Dir(uri.Path)

	return Downloader{
		client:            &http.Client{},
		masterManifestSrc: masterManifestSrc,
		baseManifestUrl:   uri,
	}
}

func (d *Downloader) SetRenditions(renditions []Rendition) {
	d.renditions = renditions
}

func (d *Downloader) DownloadMastermanifest() ([]Rendition, *RequestTraceResult, error) {
	payload, trace, err := d.downloadFile(d.masterManifestSrc)
	if err != nil {
		return []Rendition{}, trace, err
	}

	renditions, err := parseMasterManifestRenditions(string(payload))
	if err != nil {
		return []Rendition{}, trace, err

	}
	return renditions, trace, nil
}

func (d *Downloader) DownloadPlaylistByName(name string) (*Playlist, *RequestTraceResult, error) {
	rendition := d.findRendition(name)
	if rendition == nil {
		return nil, nil, errors.New("Rendition with given name has not been found")
	}

	uri := *d.baseManifestUrl
	uri.Path = path.Join(uri.Path, rendition.URL)

	return d.downloadRenditionPlaylist(uri.String())
}

func (d *Downloader) RefreshPlaylist(playlist *Playlist) (*Playlist, *RequestTraceResult, error) {
	return d.downloadRenditionPlaylist(playlist.URL)
}

func (d *Downloader) DownloadMedia(playlist *Playlist, media Media) ([]byte, *RequestTraceResult, error) {
	uri := urlDir(playlist.URL) + "/" + media.Url()

	return d.downloadFile(uri)
}

func (d *Downloader) RefreshPlaylistWithBlockingReload(playlist *Playlist) (*Playlist, *RequestTraceResult, error) {
	var lastPartSegment *PartSegment
	var foundIdx int

	for idx, item := range playlist.Items {
		if part, ok := item.(*PartSegment); ok {
			lastPartSegment = part
			foundIdx = idx
		}
	}

	var uri string
	var msn int
	var partSn int
	if foundIdx == len(playlist.Items)-1 {
		msn = lastPartSegment.SeqNum
		partSn = lastPartSegment.PartNum + 1
	} else {
		// the last element was a segment therefore ask for a next msn
		msn = lastPartSegment.SeqNum + 1
		partSn = 0
	}
	uri = fmt.Sprintf("%s?_HLS_msn=%d&_HLS_part=%d", playlist.URL, msn, partSn)

	return d.downloadRenditionPlaylist(uri)
}

func (d *Downloader) downloadRenditionPlaylist(url string) (*Playlist, *RequestTraceResult, error) {
	payload, trace, err := d.downloadFile(url)
	if err != nil {
		return nil, trace, err
	}

	playlist, err := parseRenditionManifest(string(payload))
	if err != nil {
		return nil, trace, err
	}

	idx := strings.Index(url, "?_HLS_msn")
	if idx > 0 {
		url = url[0:idx]
	}

	playlist.URL = url

	return playlist, trace, nil
}

func (d *Downloader) findRendition(name string) *Rendition {
	var targetRendition *Rendition
	for _, rendition := range d.renditions {
		if rendition.Name == name {
			targetRendition = &rendition
			break
		}
	}

	return targetRendition
}

func (d *Downloader) downloadFile(url string) ([]byte, *RequestTraceResult, error) {
	req, _ := http.NewRequest("GET", url, nil)

	trace := newRequestTrace(req)

	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace.GetClientTrace()))

	trace.SetRequestStart()
	if resp, err := d.client.Do(req); err != nil {
		return []byte{}, nil, err
	} else {
		// read response body
		defer resp.Body.Close()

		trace.SetStatus(resp.StatusCode)

		// read response body
		if body, err := io.ReadAll(resp.Body); err != nil {
			return []byte{}, nil, err
		} else {
			trace.SetDownloaded()

			if resp.StatusCode != 200 {
				return []byte{}, trace.ToRequestTraceResult(), errors.New("Invalid response status code")
			} else {
				return body, trace.ToRequestTraceResult(), nil
			}
		}
	}
}

func urlDir(url string) string {
	idx := strings.LastIndex(url, "/")
	return url[0:idx]
}
