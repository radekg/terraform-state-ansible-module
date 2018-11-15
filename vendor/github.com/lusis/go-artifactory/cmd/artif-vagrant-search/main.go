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
	criteria = kingpin.Arg("criteria", "what to search for").Required().String()
)

func main() {
	kingpin.Parse()
	output, _ := outputter.NewOutputter(*format)
	client, clientErr := artifactory.NewClientFromEnv()
	if clientErr != nil {
		fmt.Printf("%s\n", clientErr.Error())
		os.Exit(1)
	}
	data, err := client.VagrantSearch(*criteria)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	}
	theaders := []string{
		"NAME",
		"VERSION",
		"PROVIDER",
		"MODIFIED",
		"MODIFIED BY",
	}
	output.SetHeaders(theaders)

	for _, d := range data {
		props := make(map[string]string)
		for _, prop := range d.Properties {
			props[prop.Key] = prop.Value
		}
		tdata := []string{
			fmt.Sprintf("%s/%s", d.Repo, props["box_name"]),
			props["box_version"],
			props["box_provider"],
			d.Modified,
			d.ModifiedBy,
		}
		_ = output.AddRow(tdata)
	}
	output.Draw()
	os.Exit(0)
}
