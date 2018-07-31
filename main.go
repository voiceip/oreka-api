package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"encoding/xml"
	"github.com/fatih/color"

	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"fmt"
	"runtime"
	"bufio"
	"io/ioutil"
	"strings"
)

type HibernateConfiguration struct {
	XMLName xml.Name `xml:"hibernate-configuration"`
	Text    string   `xml:",chardata"`
	SessionFactory struct {
		Text string `xml:",chardata"`
		Property []struct {
			Text string `xml:",chardata"`
			Name string `xml:"name,attr"`
		} `xml:"property"`
	} `xml:"session-factory"`
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func die(msg string, e error) {
	if runtime.GOOS == "windows" {
		fmt.Println("ERROR:", msg)
	} else {
		fmt.Println(color.RedString(msg))
	}
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')
	check(e)
	os.Exit(1)
}

var configFilePath = "/etc/oreka/database.hbm.xml"
var DB *sql.DB

func setupDatabase() {

	xmlFile, err := os.Open(configFilePath)
	// if we os.Open returns an error then handle it
	check(err)
	defer xmlFile.Close()

	// read our opened xmlFile as a byte array.
	byteValue, _ := ioutil.ReadAll(xmlFile)
	var config HibernateConfiguration
	xml.Unmarshal(byteValue, &config)

	var username, password, host, database string

	for _, item := range config.SessionFactory.Property {

		if item.Name == "hibernate.connection.username" {
			username = item.Text
		} else if item.Name == "hibernate.connection.password" {
			password = item.Text
		} else if item.Name == "hibernate.connection.url" {
			parts := strings.Split(item.Text, ":")

			if parts[1] != "mysql" {
				die("Only MySQL Supported", nil)
			}
			hostdb := strings.TrimLeft(parts[2], "/")
			hostdbParts := strings.Split(hostdb, "/")

			host = hostdbParts[0]
			database = hostdbParts[1]

		}

	}

	DB, err = sql.Open("mysql", username+":"+password+"@tcp("+host+":3306)/"+database)
	check(err)
}

type OrkTape struct {
	ID              int         `json:"-"`
	Direction       int         `json:"-"`
	Duration        int         `json:"duration"`
	ExpiryTimestamp string      `json:"expiryTimestamp"`
	Filename        string      `json:"filename"`
	LocalEntryPoint string      `json:"localEntryPoint"`
	LocalParty      string      `json:"localParty"`
	PortName        string      `json:"portName"`
	RemoteParty     string      `json:"remoteParty"`
	Timestamp       string      `json:"timestamp"`
	NativeCallID    string      `json:"nativeCallId"`
	PortID          interface{} `json:"-"`
	ServiceID       int         `json:"-"`
}

func getByCallId(callId string) (OrkTape, error) {
	var tape OrkTape
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
		} else if tape != (OrkTape{}) {
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

		tape, err := getByCallId(callId)
		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal Error"})
		} else if tape != (OrkTape{}) {
			c.File("/var/log/orkaudio/audio/"+ tape.Filename)
		} else {
			c.JSON(http.StatusNotFound, gin.H{"user":user, "message": "Not Found", "callId": callId})
		}

	})

	return r
}

func main() {

	setupDatabase()
	r := setupRouter()
	r.Run(":9090")
}
