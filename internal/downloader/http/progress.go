package httpdownload

import (
	"io"
	"sync"

	"github.com/schollz/progressbar/v3"
)

type Progress struct {
	bar *progressbar.ProgressBar
	mu  sync.Mutex
}

func NewProgress(total int64) *Progress {
	bar := progressbar.NewOptions64(
		total,
		progressbar.OptionSetDescription("Downloading"),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(30),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)
	return &Progress{bar: bar}
}

func (p *Progress) Add(n int64) {
	p.mu.Lock()
	p.bar.Add64(n)
	p.mu.Unlock()
}

func (p *Progress) Close() {
	p.bar.Close()
}

type ProgressWriter struct {
	progress *Progress
	written  int64
}

func NewProgressWriter(p *Progress) *ProgressWriter {
	return &ProgressWriter{progress: p}
}

func (pw *ProgressWriter) Write(b []byte) (int, error) {
	n := len(b)
	pw.written += int64(n)
	pw.progress.Add(int64(n))
	return n, nil
}

func (pw *ProgressWriter) Reset() {
	if pw.written > 0 {
		pw.progress.Add(-pw.written)
		pw.written = 0
	}
}

var _ io.Writer = (*ProgressWriter)(nil)
