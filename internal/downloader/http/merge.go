package httpdownload

import (
	"fmt"
	"io"
	"os"
)

func MergeSegments(segments []*Segment, outputPath string) error {
	out, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create output: %w", err)
	}
	defer out.Close()

	for _, seg := range segments {
		if err := appendFile(out, seg.TempFile); err != nil {
			return fmt.Errorf("append segment %d: %w", seg.Index, err)
		}
	}
	return nil
}

func appendFile(dst *os.File, srcPath string) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	_, err = io.Copy(dst, src)
	return err
}

func CleanupSegments(segments []*Segment) {
	for _, seg := range segments {
		seg.Cleanup()
	}
}
