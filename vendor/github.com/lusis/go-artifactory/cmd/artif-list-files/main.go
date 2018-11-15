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
	repo = kingpin.Arg("repo", "repo to list files").Required().String()
	path = kingpin.Arg("path", "path to list files").Default("/").String()
)

func main() {
	kingpin.Parse()
	output, _ := outputter.NewOutputter(*format)
	client, clientErr := artifactory.NewClientFromEnv()
	if clientErr != nil {
		fmt.Printf("%s\n", clientErr.Error())
		os.Exit(1)
	}
	u, err := client.GetFileList(*repo, *path)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	} else {
		output.SetHeaders([]string{"URI", "Size", "SHA-1"})
		for _, v := range u.Files {
			_ = output.AddRow([]string{v.URI, fmt.Sprintf("%d", v.Size), v.SHA1})
		}

		output.Draw()
		os.Exit(0)
	}
}
