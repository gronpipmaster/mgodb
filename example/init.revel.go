/*
Example init dbm for Revel framework
*/
package controllers

import (
	"fmt"
	"github.com/gronpipmaster/mgodb"
	"github.com/robfig/revel"
	"net/url"
	"time"
)

func init() {
	revel.OnAppStart(AppStart)
}

func AppStart() {
	var dbm *mgodb.Dbm
	var err error
	connectUrl, dbName, timeout := getConnectUrlDb()
	err = dbm.Init(connectUrl, dbName, timeout)
	if err != nil {
		revel.ERROR.Fatal(err)
	}
}

func getConnectUrlDb() (string, string, time.Duration) {
	var err error
	var found bool
	var host, port, user, pass, dbName, timeoutStr string
	var timeout time.Duration

	connURL := &url.URL{Scheme: "mongodb"}

	if host, found = revel.Config.String("mongo.host"); !found {
		revel.ERROR.Fatal("No mongo.host found.")
	}
	if port, found = revel.Config.String("mongo.port"); !found {
		revel.ERROR.Fatal("No mongo.port found.")
	}
	if user, found = revel.Config.String("mongo.user"); !found {
		revel.ERROR.Fatal("No mongo.user found.")
	}
	if pass, found = revel.Config.String("mongo.pass"); !found {
		revel.ERROR.Fatal("No mongo.pass found.")
	}
	if dbName, found = revel.Config.String("mongo.db"); !found {
		revel.ERROR.Fatal("No mongo.db found.")
	}
	if timeoutStr, found = revel.Config.String("mongo.timeout"); !found {
		timeoutStr = "5s"
	}
	timeout, err = time.ParseDuration(timeoutStr)
	if err != nil {
		revel.ERROR.Fatal("mongo.timeout found error, ", err.Error())
	}

	connURL.Host = fmt.Sprintf("%s:%s", host, port)
	connURL.User = url.UserPassword(user, pass)
	connURL.Path = "/" + dbName
	return connURL.String(), dbName, timeout
}
