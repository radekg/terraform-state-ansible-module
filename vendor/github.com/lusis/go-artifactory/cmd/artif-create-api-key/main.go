package main

import (
	"fmt"
	"os"

	artifactory "github.com/lusis/go-artifactory/artifactory.v54"
)

func main() {
	client, clientErr := artifactory.NewClientFromEnv()
	if clientErr != nil {
		fmt.Printf("%s\n", clientErr.Error())

	} else {
		p, err := client.CreateUserAPIKey()
		if err != nil {
			fmt.Printf("%s\n", err)
			os.Exit(1)
		} else {
			fmt.Printf("%s\n", p)
			os.Exit(0)
		}
	}
}
