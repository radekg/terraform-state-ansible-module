package main

import (
	"fmt"
	"os"

	artifactory "github.com/lusis/go-artifactory/artifactory.v54"
	"github.com/lusis/outputter"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	formatUsage = fmt.Sprintf("Format to show results [table, csv, list (usernames only - useful for piping)]")
	format      = kingpin.Flag("format", "format to display output").
			Default("table").
			Enum(outputter.GetOutputters()...)
)

func main() {
	kingpin.UsageTemplate(kingpin.CompactUsageTemplate).Version("1.0").Author("John E. Vincent")
	kingpin.CommandLine.Help = "List all users in Artifactory"
	kingpin.Parse()
	output, _ := outputter.NewOutputter(*format)
	client, clientErr := artifactory.NewClientFromEnv()
	if clientErr != nil {
		fmt.Printf("%s\n", clientErr.Error())
		os.Exit(1)
	}
	data, err := client.GetUsers()
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
