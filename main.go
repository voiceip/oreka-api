package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"fmt"
	"./oreka"
	"github.com/viert/lame"

	"os"
	"bufio"
)

var DB *sql.DB
var err error

func getByCallId(callId string) (oreka.OrkTape, error) {
	var tape oreka.OrkTape
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
		if count > 0 {
			return tape, nil
		} else {
			return tape, nil
		}
	}

}

func setupRouter() *gin.Engine {
	// Disable Console Color
	// gin.DisableConsoleColor()
	r := gin.Default()

	// Ping test
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	r.GET("/calls/:id", func(c *gin.Context) {
		callId := c.Params.ByName("id")

		tape, err := getByCallId(callId)
		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Error"})
		} else if tape != (oreka.OrkTape{}) {
			c.JSON(http.StatusOK, tape)
		} else {
			c.JSON(http.StatusNotFound, gin.H{"message": "Not Found", "callId": callId})

		}

	})

	authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
		"root":  "root",
		"admin": "admin",
	}))

	authorized.GET("/play/:id", func(c *gin.Context) {
		user := c.MustGet(gin.AuthUserKey).(string)
		callId := c.Params.ByName("id")
		format := c.DefaultQuery("format", "wav")

		tape, err := getByCallId(callId)
		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Error"})
		} else if tape != (oreka.OrkTape{}) {
			wavSourceFile := "/var/log/orkaudio/audio/" + tape.Filename
			switch format {
			case "mp3":
				f, err := os.Open(wavSourceFile)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Error", "debug": err.Error()})
				} else {
					defer f.Close()
					reader := bufio.NewReader(f)

					wr := lame.NewWriter(c.Writer)
					wr.Encoder.SetBitrate(112)
					wr.Encoder.SetQuality(1)
					wr.Encoder.InitParams()

					//c.Header("Transfer-Encoding", "chunked")
					c.Header("Content-Disposition", `inline; filename="`+callId+`.mp3"`)
					c.Header("Content-Type", "audio/mp3")
					reader.WriteTo(wr)

				}

			default:
				c.File(wavSourceFile)
			}
		} else {
			c.JSON(http.StatusNotFound, gin.H{"user": user, "message": "Not Found", "callId": callId})
		}

	})

	return r
}

func main() {

	DB, err = oreka.SetupDatabase()
	if err != nil {
		oreka.Die("Unable to Connect to Database", err)
	}
	r := setupRouter()
	r.Run(":9090")
}
