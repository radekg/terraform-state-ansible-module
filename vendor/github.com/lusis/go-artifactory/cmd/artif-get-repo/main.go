package main

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/lusis/outputter"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	artifactory "github.com/lusis/go-artifactory/artifactory.v54"
)

var (
	format = kingpin.Flag("format", "format to display output").
		Default("table").
		Enum("table", "json", "tabular")
	repo = kingpin.Arg("repo", "Repo to show").Required().String()
)

func makeBaseRow(b artifactory.RepoConfig) []string {
	s := reflect.ValueOf(b)
	baseRow := []string{
		s.FieldByName("Key").String(),
		s.FieldByName("RClass").String(),
		s.FieldByName("PackageType").String(),
		s.FieldByName("Description").String(),
		s.FieldByName("Notes").String(),
		strconv.FormatBool((s.FieldByName("BlackedOut").Bool())),
		strconv.FormatBool((s.FieldByName("HandleReleases").Bool())),
		strconv.FormatBool((s.FieldByName("HandleSnapshots").Bool())),
		s.FieldByName("ExcludesPattern").String(),
		s.FieldByName("IncludesPattern").String(),
	}
	return baseRow
}

func main() {
	kingpin.Parse()
	output, _ := outputter.NewOutputter(*format)
	client, clientErr := artifactory.NewClientFromEnv()
	if clientErr != nil {
		fmt.Printf("%s\n", clientErr.Error())
		os.Exit(1)
	}
	data, err := client.GetRepo(*repo, make(map[string]string))
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	} else {
		baseHeaders := []string{
			"Key",
			"Type",
			"PackageType",
			"Description",
			"Notes",
			"Blacked Out?",
			"Releases?",
			"Snapshots?",
			"Excludes",
			"Includes",
		}

		// base row data common to all repos
		baseRow := makeBaseRow(data)
		// We have to do this to get to the concrete repo type
		switch data.MimeType() {
		case artifactory.RemoteRepoMimeType:
			d := data.(artifactory.RemoteRepoConfig)
			baseHeaders = append(baseHeaders, "Url")
			output.SetHeaders(baseHeaders)
			baseRow = append(baseRow, d.URL)
			_ = output.AddRow(baseRow)
		case artifactory.LocalRepoMimeType:
			d := data.(artifactory.LocalRepoConfig)
			baseHeaders = append(baseHeaders, "Layout")
			baseRow = append(baseRow, d.LayoutRef)
			output.SetHeaders(baseHeaders)
			_ = output.AddRow(baseRow)
		case artifactory.VirtualRepoMimeType:
			d := data.(artifactory.VirtualRepoConfig)
			baseHeaders = append(baseHeaders, "Repositories")
			baseRow = append(baseRow, strings.Join(d.Repositories, "\n"))
			output.SetHeaders(baseHeaders)
			_ = output.AddRow(baseRow)
		default:
			output.SetHeaders(baseHeaders)
			_ = output.AddRow(baseRow)
		}
		output.Draw()
		os.Exit(0)
	}
}
