package main

import (
	"fmt"
	"os"

	artifactory "github.com/lusis/go-artifactory/artifactory.v54"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	username = kingpin.Arg("username", "Username to delete").Required().String()
)

func main() {
	kingpin.Parse()
	client, clientErr := artifactory.NewClientFromEnv()
	if clientErr != nil {
		fmt.Printf("%s\n", clientErr.Error())
		os.Exit(1)
	}
	err := client.DeleteUser(*username)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	} else {
		fmt.Printf("User %s deleted\n", *username)
		os.Exit(0)
	}
}
