package oreka

import (
	"fmt"
	"os/exec"
	"os"
	"github.com/go-audio/wav"
	"github.com/go-audio/audio"
	"io"
)

type MediaInfo struct {
	Length      int64
	ContentType string
	Data        io.ReadCloser
}

type WAVProcessor struct{
	WAVFile string
	ID *string
	Size *int
	Duration *int

}
var tmpFilePath = `/tmp/%s.mp3`

func (mp *WAVProcessor) ToMP3() (*deleteCloser, error) {
	mp3FileName := fmt.Sprintf(tmpFilePath,mp.ID)
	ffmpegResult, err := exec.Command("ffmpeg", "-y", "-i" , mp.WAVFile, mp3FileName ).CombinedOutput()
	if err != nil {
		fmt.Println("FFMPEG Error" , string(ffmpegResult))
		return nil, err
	}
	return DeleteOnCloseReader(mp3FileName)
}

func (mp *WAVProcessor) HasSilentStream() (bool, error) {
	f, err := os.Open(mp.WAVFile)
	if err != nil {
		return  false, err
	}
	defer f.Close()

	decoder := wav.NewDecoder(f)

	if !decoder.IsValidFile(){
		return false,  fmt.Errorf("NOT_VALID_MEDIA")
	}

	format := &audio.Format{
		NumChannels: int(decoder.NumChans),
		SampleRate:  int(decoder.SampleRate),
	}

	chunkSize := 4096
	buf := &audio.IntBuffer{Data: make([]int, chunkSize), Format: format}
	var n int

	var channelProps = make([]bool, buf.Format.NumChannels)
	for i := 0; i < buf.Format.NumChannels; i++ {
		channelProps[i] = true
	}

	for err == nil {
		n, err = decoder.PCMBuffer(buf)
		if err != nil {
			return false, err
		}
		if n == 0 {
			break
		}
		for i, s := range buf.Data {
			// the buffer is longer than than the data we have, we are done
			if i == n {
				break
			}

			if s != 255 {
				channelProps[i % buf.Format.NumChannels] = false
			}
		}
		if n != chunkSize {
			break
		}
	}

	for i := 0; i < buf.Format.NumChannels; i++ {
		if channelProps[i] == true{
			return true, nil
		}
	}

	return false, nil

}
