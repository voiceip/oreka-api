package main

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/voiceip/oreka-api/oreka"
)

var DB *sql.DB
var err error

const recordBasePath = "/var/log/orkaudio/audio/"

func getByCallID(callID string) (*oreka.OrkTape, error) {
	var tape oreka.OrkTape
	if callID == "" {
		return nil, nil
	}
	results, err := DB.Query("select filename, duration,  localParty, remoteParty, timestamp, nativeCallId, state from orktape where `nativeCallId` = ?", callID)
	if err != nil {
		return nil, err
	}
	defer results.Close()
	for results.Next() {
		err := results.Scan(&tape.Filename, &tape.Duration, &tape.LocalParty, &tape.RemoteParty, &tape.Timestamp, &tape.NativeCallID, &tape.CallState)
		if err != nil {
			return nil, err
		}
		return &tape, nil
	}
	return nil, nil
}

func setupRouter() *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()

	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		"root":  "root",
		"admin": "admin",
	}))

	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	r.GET("/calls/:id", func(c *gin.Context) {
		callId := c.Params.ByName("id")
		tape, err := getByCallID(callId)
		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Error", "reason": err.Error()})
		} else if tape != nil {
			if _, err := os.Stat(recordBasePath + tape.Filename); os.IsNotExist(err) {
				c.JSON(http.StatusGone, tape)
			} else {
				c.JSON(http.StatusOK, tape)
			}
		} else {
			c.JSON(http.StatusNotFound, gin.H{"message": "Not Found", "callId": callId})
		}
	})

	authorized.GET("/play/:id", func(c *gin.Context) {
		user := c.MustGet(gin.AuthUserKey).(string)
		callID := c.Params.ByName("id")
		format := c.DefaultQuery("format", "")

		tape, err := getByCallID(callID)
		if err == nil && tape != nil {
			err = convertMediaFileIfRequired(tape)
		}
		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Error", "reason": err.Error()})
		} else if tape != nil {
			sourceMediaFile := recordBasePath + tape.Filename
			serveMediaFile(c, callID, format, sourceMediaFile)
		} else {
			c.JSON(http.StatusNotFound, gin.H{"user": user, "message": "Not Found", "callId": callID})
		}
	})

	return r
}

func convertMediaFileIfRequired(tape *oreka.OrkTape) error {
	sourceMediaFile := recordBasePath + tape.Filename
	_, err := os.Stat(sourceMediaFile)
	if os.IsNotExist(err) && tape.CallState == "STOP" {
		//check if mcfFile is available
		mcfMediaFile := sourceMediaFile[0:strings.LastIndex(sourceMediaFile, ".")] + ".mcf"
		if _, err := os.Stat(mcfMediaFile); err == nil {
			fmt.Printf("Running Transcode on NativeCallID %s, MCF : %s \n", tape.NativeCallID, mcfMediaFile)
			err = oreka.OrkaudioTranscode(mcfMediaFile)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func serveMediaFile(c *gin.Context, callId string, format string, mediaFile string) {

	if _, err := os.Stat(mediaFile); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"message": "Raw file not found.", "callId": callId})
	} else {
		fileExtension := strings.TrimLeft(path.Ext(mediaFile), ".")
		switch format {
		case "mp3":
			if fileExtension == format {
				c.File(mediaFile)
			} else {
				mp := oreka.MediaProcessor{FileName: mediaFile, ID: &callId}
				stream, err := mp.ToMP3()
				defer stream.Close()
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Error", "error": err.Error()})
				} else {
					//c.Header("Transfer-Encoding", "chunked")
					c.Header("Content-Length", strconv.FormatInt(stream.Size(), 10))
					c.Header("Content-Disposition", `inline; filename="`+callId+`.mp3"`)
					c.Header("Content-Type", "audio/mp3")
					io.Copy(c.Writer, stream)
				}
			}
		case "wav", "ogg", "opus":
			if fileExtension == format {
				c.File(mediaFile)
			} else {
				c.JSON(http.StatusUnsupportedMediaType, gin.H{"message": "Unsupported Format"})
			}
		default:
			c.File(mediaFile)
		}
	}
}

func main() {

	DB, err = oreka.SetupDatabase()
	if err != nil {
		oreka.Die("Unable to Connect to Database", err)
	}
	r := setupRouter()
	r.Run(":9090")
}
