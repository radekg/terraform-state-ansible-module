package main

import (
	"fmt"
	"os"
	"strings"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	artifactory "github.com/lusis/go-artifactory/artifactory.v54"
	"github.com/lusis/outputter"
)

var (
	format = kingpin.Flag("format", "format to display output").
		Default("table").
		Enum(outputter.GetOutputters()...)
	groupid    = kingpin.Flag("groupid", "groupid coordinate").String()
	artifactid = kingpin.Flag("artifactid", "artifactid coordinate").String()
	version    = kingpin.Flag("version", "version coordinate").String()
	classifier = kingpin.Flag("classifier", "classifier coordinate").String()
	repo       = kingpin.Flag("repo", "repo to search against. can be specified multiple times").Strings()
)

func main() {
	kingpin.Parse()
	output, _ := outputter.NewOutputter(*format)
	client, clientErr := artifactory.NewClientFromEnv()
	if clientErr != nil {
		fmt.Printf("%s\n", clientErr.Error())
		os.Exit(1)
	}
	var coords artifactory.GAVC
	if groupid != nil {
		coords.GroupID = *groupid
	}
	if artifactid != nil {
		coords.ArtifactID = *artifactid
	}
	if version != nil {
		coords.Version = *version
	}
	if classifier != nil {
		coords.Classifier = *classifier
	}
	if repo != nil {
		coords.Repos = *repo
	}
	data, err := client.GAVCSearch(&coords)
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	} else {
		output.SetHeaders([]string{
			"File",
			"Repo",
			"RemoteUrl",
			"Created",
			"Last Modified",
			"Created By",
			"Modified By",
			"SHA1",
			"MD5",
			"Size",
			"MimeType",
			"Download",
		})
		for _, r := range data {
			elems := strings.Split(r.Path, "/")
			fileName := elems[len(elems)-1]
			_ = output.AddRow([]string{
				fileName,
				r.Repo,
				r.RemoteURL,
				r.Created,
				r.LastModified,
				r.CreatedBy,
				r.ModifiedBy,
				r.Checksums.SHA1,
				r.Checksums.MD5,
				r.Size,
				r.MimeType,
				r.DownloadURI,
			})
		}
		output.Draw()
		os.Exit(0)
	}
}
