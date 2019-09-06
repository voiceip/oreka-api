package oreka

import (
	"fmt"
	"os/exec"
)

type MediaProcessor struct{
	WAVFile string
	ID string
	Size int

}
var tmpFilePath = `/tmp/%s.mp3`

func (mp *MediaProcessor) ToMP3() (*deleteCloser, error) {
	mp3FileName := fmt.Sprintf(tmpFilePath,mp.ID)
	ffmpegResult, err := exec.Command("ffmpeg", "-y", "-i" , mp.WAVFile, mp3FileName ).CombinedOutput()
	if err != nil {
		fmt.Println("FFMPEG Error" , string(ffmpegResult))
		return nil, err
	}
	return DeleteOnCloseReader(mp3FileName)
}
