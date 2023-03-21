package hls

import (
	"errors"
	"github.com/Qizot/go-m3u8/m3u8"
	"time"
)

func parseMasterManifestRenditions(payload string) ([]Rendition, error) {

	var renditions []Rendition = make([]Rendition, 0, 10)
	manifest, err := m3u8.ReadString(payload)
	if err != nil {
		return []Rendition{}, err
	}

	for _, item := range manifest.Items {
		playlistItem, ok := item.(*m3u8.PlaylistItem)
		if !ok {
			return []Rendition{}, errors.New("Missing playlist item")
		}
		renditions = append(renditions, Rendition{
			Bandwidth: playlistItem.Bandwidth,
			Name:      *playlistItem.Name,
			URL:       playlistItem.URI,
		})
	}

	return renditions, nil
}

func parseRenditionManifest(payload string) (*Playlist, error) {
	manifest, err := m3u8.ReadString(payload)
	if err != nil {
		return nil, err
	}

	playlist := &Playlist{
		TargetDuration:     manifest.Target,
		PartTargetDuration: manifest.PartTarget,
		PartHoldBack:       manifest.PartHoldBack,
	}

	msn := manifest.Sequence
	partSn := 0
	var latestProgramDateTime time.Time

	for _, item := range manifest.Items {
		switch v := item.(type) {
		case *m3u8.TimeItem:
			latestProgramDateTime = v.Time
		case *m3u8.SegmentItem:
			segment := &Segment{
				Duration: v.Duration,
				Uri:      v.Segment,
				SequenceNumbers: SequenceNumbers{
					SeqNum:  msn,
					PartNum: -1,
				},
			}

			if !latestProgramDateTime.IsZero() {
				segment.ProgramDateTime = latestProgramDateTime
				latestProgramDateTime = time.Time{}
			}

			playlist.Items = append(playlist.Items, segment)

			msn += 1
			partSn = 0

		case *m3u8.PartSegmentItem:
			partSegment := &PartSegment{
				Duration:    v.Duration,
				Uri:         v.Uri,
				Independent: v.Independent,
				SequenceNumbers: SequenceNumbers{
					SeqNum:  msn,
					PartNum: partSn,
				},
			}

			if !latestProgramDateTime.IsZero() {
				partSegment.ProgramDateTime = &latestProgramDateTime
				latestProgramDateTime = time.Time{}
			}

			playlist.Items = append(playlist.Items, partSegment)

			partSn += 1

		default:
		}
	}

	return playlist, nil
}
