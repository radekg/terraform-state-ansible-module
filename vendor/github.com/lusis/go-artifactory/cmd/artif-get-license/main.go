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
		Enum("table", "json", "tabular")
)

func main() {
	kingpin.Parse()
	output, _ := outputter.NewOutputter(*format)
	client, clientErr := artifactory.NewClientFromEnv()
	if clientErr != nil {
		fmt.Printf("%s\n", clientErr.Error())
		os.Exit(1)
	}
	data, err := client.GetLicenseInfo()
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	} else {
		output.SetHeaders([]string{"Type", "Expires", "Owner"})
		_ = output.AddRow([]string{data.LicenseType, data.ValidThrough, data.LicensedTo})
		output.Draw()
		os.Exit(0)
	}
}
