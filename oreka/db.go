package oreka

import (
	"database/sql"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

var configFilePath = "/etc/oreka/database.hbm.xml"

type HibernateConfiguration struct {
	XMLName        xml.Name `xml:"hibernate-configuration"`
	Text           string   `xml:",chardata"`
	SessionFactory struct {
		Text     string `xml:",chardata"`
		Property []struct {
			Text string `xml:",chardata"`
			Name string `xml:"name,attr"`
		} `xml:"property"`
	} `xml:"session-factory"`
}

func SetupDatabase() (*sql.DB, error) {

	xmlFile, err := os.Open(configFilePath)
	// if we os.Open returns an error then handle it

	if err != nil {
		return nil, err
	}

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
				return nil, fmt.Errorf("Only MySQL Supported")
			}
			hostdb := strings.TrimLeft(parts[2], "/")
			hostdbParts := strings.Split(hostdb, "/")

			host = hostdbParts[0]
			database = hostdbParts[1]
			//remove params
			if pos := strings.Index(database, "?"); pos >= 0 {
				database = database[:pos]
			}

		}
	}

	DB, err := sql.Open("mysql", username+":"+password+"@tcp("+host+":3306)/"+database)

	if err != nil {
		return nil, err
	}

	return DB, nil
}
