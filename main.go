package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"fmt"
	"github.com/voiceip/oreka-api/oreka"
	"io"
	"strconv"
	"path"
	"strings"
	"os"
)

var DB *sql.DB
var err error

const recordBasePath = "/var/log/orkaudio/audio/"

func getByCallId(callId string) (oreka.OrkTape, error) {
	var tape oreka.OrkTape
	if callId == "" {
		return tape, nil
	}
	results, err := DB.Query("select filename, duration,  localParty, remoteParty, timestamp, nativeCallId from orktape where `nativeCallId` = ?", callId)
	if err != nil {
		return tape, err
	} else {
		count := 0
		for results.Next() {
			count += 1
			// for each row, scan the result into our tag composite object
			err := results.Scan(&tape.Filename, &tape.Duration, &tape.LocalParty, &tape.RemoteParty, &tape.Timestamp, &tape.NativeCallID)
			if err != nil {
				return tape, err
			}
		}
		return tape, nil
	}

}

func setupRouter() *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()

	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		"root": "root",
		"admin": "admin",
	}))

	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	r.GET("/calls/:id", func(c *gin.Context) {
		callId := c.Params.ByName("id")
		tape, err := getByCallId(callId)
		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Error", "reason" :  err.Error()})
		} else if tape != (oreka.OrkTape{}) {
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
		callId := c.Params.ByName("id")
		format := c.DefaultQuery("format", "")

		tape, err := getByCallId(callId)
		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Error", "reason" :  err.Error()})
		} else if tape != (oreka.OrkTape{}) {
			sourceMediaFile := recordBasePath + tape.Filename
			serveMediaFile(c, callId, format, sourceMediaFile)
		} else {
			c.JSON(http.StatusNotFound, gin.H{"user": user, "message": "Not Found", "callId": callId})
		}
	})

	return r
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
		case "wav", "ogg", "opus" :
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
