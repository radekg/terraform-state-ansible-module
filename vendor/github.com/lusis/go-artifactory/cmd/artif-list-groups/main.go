package main

import (
	"fmt"
	"os"

	artifactory "github.com/lusis/go-artifactory/artifactory.v54"
	"github.com/lusis/outputter"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	format = kingpin.Flag("format", "format to display output").
		Default("table").
		Enum(outputter.GetOutputters()...)
)

func main() {
	kingpin.Parse()
	client, clientErr := artifactory.NewClientFromEnv()
	if clientErr != nil {
		fmt.Printf("%s\n", clientErr.Error())
		os.Exit(1)
	}
	output, _ := outputter.NewOutputter(*format)
	data, err := client.GetGroups()
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	} else {
		output.SetHeaders([]string{"Name", "Uri"})
		for _, u := range data {
			_ = output.AddRow([]string{u.Name, u.URI})
		}
		output.Draw()
		os.Exit(0)
	}
}
