package hls

import "time"

type Rendition struct {
	Bandwidth int
	URL       string
	Name      string
}

type SequenceNumbers struct {
	SeqNum  int
	PartNum int
}

type Media interface {
	Url() string
}

type Playlist struct {
	URL                string
	TargetDuration     int
	PartTargetDuration float64
	PartHoldBack       float64
	Items              []Media
}

func (p *Playlist) LastMediaItem() Media {
	return p.Items[len(p.Items)-1]
}

func (p *Playlist) NextSeqNum() SequenceNumbers {
	switch item := p.LastMediaItem().(type) {
	case *Segment:
		return SequenceNumbers{SeqNum: item.SeqNum + 1, PartNum: 0}
	case *PartSegment:
		return SequenceNumbers{SeqNum: item.SeqNum, PartNum: item.PartNum + 1}
	}

	return SequenceNumbers{SeqNum: 0, PartNum: 0}
}

type Segment struct {
	SequenceNumbers
	Duration        float64
	Uri             string
	ProgramDateTime time.Time
}

func (s *Segment) Url() string {
	return s.Uri
}

type PartSegment struct {
	SequenceNumbers
	Duration        float64
	Uri             string
	ProgramDateTime *time.Time
	Independent     bool
}

func (s *PartSegment) Url() string {
	return s.Uri
}

func (s *SequenceNumbers) GreaterOrEqual(other *SequenceNumbers) bool {
	return s.SeqNum >= other.SeqNum && s.PartNum >= other.PartNum
}
