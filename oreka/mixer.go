package oreka

import (
	"fmt"
	"os/exec"
	"os"
)

var mixMP3FilePath = `/tmp/%s.mp3`

/**
	Will only mix if individual streams are mono or contains a silent stream
 */
func MixMediaStreams(callId string, mediaTrackA string, mediaTrackB string) (*MediaInfo, error) {
	mediaInfoTrackA := &WAVProcessor{WAVFile:mediaTrackA}
	hasSilence, err := mediaInfoTrackA.HasSilentStream()
	if err != nil {
		return nil, err
	}

	if !hasSilence {
		out, _ := os.Open(mediaTrackA)
		fi, _ := out.Stat()
		return &MediaInfo{
			fi.Size(),
			"audio/wav",
			out,
		}, nil
	}

	mediaInfoTrackB := &WAVProcessor{WAVFile:mediaTrackB}
	hasSilence, err = mediaInfoTrackB.HasSilentStream()
	if err != nil {
		return nil, err
	}

	if !hasSilence {
		out, _ := os.Open(mediaTrackB)
		fi, _ := out.Stat()
		return &MediaInfo{
			fi.Size(),
			"audio/wav",
			out,
		}, nil
	}

	//Here iff both tracks have silence tracks
	muxFileName := fmt.Sprintf(mixMP3FilePath, callId)

	ffmpegResult, err := exec.Command("ffmpeg", "-y", "-i", mediaTrackA, "-i", mediaTrackB, "-filter_complex", "amerge", "-ac", "2", "-c:a", "libmp3lame", "-q:a", "4", muxFileName).CombinedOutput()
	if err != nil {
		fmt.Println("FFMPEG Error", string(ffmpegResult))
		return nil, err
	}

	out, _ := DeleteOnCloseReader(muxFileName)
	return &MediaInfo{
		out.Size(),
		"audio/mp3",
		out,
	}, nil

}
