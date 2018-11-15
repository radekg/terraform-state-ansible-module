package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/lusis/outputter"

	artifactory "github.com/lusis/go-artifactory/artifactory.v54"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	format = kingpin.Flag("format", "format to display output").
		Default("table").
		Enum("table", "json", "tabular")
	group = kingpin.Arg("group", "group name to show").Required().String()
)

func main() {
	kingpin.Parse()
	output, _ := outputter.NewOutputter(*format)
	client, cErr := artifactory.NewClientFromEnv()
	if cErr != nil {
		fmt.Printf("unable to get an artifactory client: %s\n", cErr.Error())
		os.Exit(1)
	}
	u, err := client.GetGroupDetails(*group, make(map[string]string))
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	} else {
		output.SetHeaders([]string{"Name", "Description", "AutoJoin?", "Realm", "Realm Attributes"})
		_ = output.AddRow([]string{
			u.Name,
			u.Description,
			strconv.FormatBool(u.AutoJoin),
			u.Realm,
			u.RealmAttributes,
		})
		output.Draw()
		os.Exit(0)
	}
}
