package oreka

import (
	"fmt"
	"io"
	"os/exec"
)

type MediaInfo struct {
	Length      int64
	ContentType string
	Data        io.ReadCloser
}

type MediaProcessor struct {
	FileName string
	ID       *string
	Size     *int
	Duration *int
}

var tmpFilePath = `/tmp/%s.mp3`

func (mp *MediaProcessor) ToMP3() (*deleteCloser, error) {
	mp3FileName := fmt.Sprintf(tmpFilePath, mp.ID)
	ffmpegResult, err := exec.Command("ffmpeg", "-y", "-i", mp.FileName, mp3FileName).CombinedOutput()
	if err != nil {
		fmt.Println("FFMPEG Error", string(ffmpegResult))
		return nil, err
	}
	return DeleteOnCloseReader(mp3FileName)
}

// OrkaudioTranscode trigger orkadudio trancode <filename>
func OrkaudioTranscode(filename string) error {
	transcodeError, err := exec.Command("orkaudio", "transcode", filename).CombinedOutput()
	if err != nil {
		fmt.Println("OrkaudioTranscode Error", string(transcodeError))
		return err
	}
	return nil
}
