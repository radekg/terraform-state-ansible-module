package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/lusis/outputter"

	artifactory "github.com/lusis/go-artifactory/artifactory.v54"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	format = kingpin.Flag("format", "format to display output").
		Default("table").
		Enum(outputter.GetOutputters()...)
	criteria   = kingpin.Arg("criteria", "what to search for").Required().String()
	showLabels = kingpin.Flag("labels", "show labels").Bool()
)

func main() {
	kingpin.Parse()
	client, clientErr := artifactory.NewClientFromEnv()
	if clientErr != nil {
		fmt.Printf("error getting artifactory client: %s\n", clientErr.Error())
		os.Exit(1)
	}
	data, err := client.DockerSearch(*criteria)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
	output, _ := outputter.NewOutputter(*format)

	theaders := []string{
		"NAME",
		"DESCRIPTION",
		"LAST MODIFIED",
		"MODIFIED BY",
	}
	if *showLabels {
		theaders = append(theaders, "LABELS")
	}
	output.SetHeaders(theaders)

	for _, d := range data {
		var description string
		if d.Properties["docker.label.description"] != nil {
			description = d.Properties["docker.label.description"][0]
		} else {
			description = "no description"
		}
		tdata := []string{
			fmt.Sprintf("%s:%s\n", d.Properties["docker.repoName"][0], d.Properties["docker.manifest"][0]),
			description,
			d.LastModified,
			d.ModifiedBy,
		}
		if *showLabels {
			var allLabels string
			var desc []string
			for p := range d.Properties {
				var labels []string
				if strings.HasPrefix(p, "docker.label") {
					labels = append(labels, p)
				}
				for _, label := range labels {
					desc = append(desc, fmt.Sprintf("%s = %s", label, d.Properties[label][0]))
				}
				allLabels = strings.Join(desc, "\n")
			}
			tdata = append(tdata, allLabels)
		}
		_ = output.AddRow(tdata)
	}
	output.Draw()
	os.Exit(0)
}
