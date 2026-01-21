package httpdownload

import (
	"fmt"
	"os"
)

type Segment struct {
	Index    int
	Start    int64
	End      int64
	TempFile string
}

type SegmentResult struct {
	Segment *Segment
	Err     error
}

func NewSegments(filename string, contentLength int64, count int) []*Segment {
	segments := make([]*Segment, count)
	segmentSize := contentLength / int64(count)

	for i := range count {
		start := int64(i) * segmentSize
		end := start + segmentSize - 1
		if i == count-1 {
			end = contentLength - 1
		}
		segments[i] = &Segment{
			Index:    i,
			Start:    start,
			End:      end,
			TempFile: fmt.Sprintf("%s_segment_%d", filename, i),
		}
	}
	return segments
}

func (s *Segment) Cleanup() {
	os.Remove(s.TempFile)
}
