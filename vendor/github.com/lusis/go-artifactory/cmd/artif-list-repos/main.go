package main

import (
	"fmt"
	"os"

	"github.com/lusis/outputter"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	artifactory "github.com/lusis/go-artifactory/artifactory.v54"
)

var (
	format = kingpin.Flag("format", "format to display output").
		Default("table").
		Enum(outputter.GetOutputters()...)
	kind = kingpin.Flag("kind", "Types of repos to show").Default("all").Enum("local", "remote", "virtual", "all")
)

func main() {
	kingpin.Parse()
	output, _ := outputter.NewOutputter(*format)
	client, clientErr := artifactory.NewClientFromEnv()
	if clientErr != nil {
		fmt.Printf("%s\n", clientErr.Error())
		os.Exit(1)
	}
	data, err := client.GetRepos(*kind)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	} else {
		output.SetHeaders([]string{
			"Key",
			"Type",
			"Description",
			"Url",
		})
		for _, r := range data {
			_ = output.AddRow([]string{
				r.Key,
				r.Rtype,
				r.Description,
				r.URL,
			})
		}
		output.Draw()
		os.Exit(0)
	}
}
