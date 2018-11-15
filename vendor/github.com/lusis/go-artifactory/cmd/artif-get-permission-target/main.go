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
		Enum("table", "json", "tabular")
	target = kingpin.Arg("target", "permission target to show").Required().String()
)

func main() {
	kingpin.Parse()
	output, _ := outputter.NewOutputter(*format)
	client, clientErr := artifactory.NewClientFromEnv()
	if clientErr != nil {
		fmt.Printf("unable to get artifactory client: %s\n", clientErr.Error())
	}
	u, err := client.GetPermissionTargetDetails(*target, make(map[string]string))
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	} else {
		output.SetHeaders([]string{"Name", "Includes", "Excludes", "Repositories", "Users", "Groups"})
		row := []string{
			u.Name,
			u.IncludesPattern,
			u.ExcludesPattern,
			strings.Join(u.Repositories, "\n"),
		}
		var users []string
		var groups []string
		for k, v := range u.Principals.Users {
			line := fmt.Sprintf("%s (%s)", k, strings.Join(v, ","))
			users = append(users, line)
		}
		for k, v := range u.Principals.Groups {
			line := fmt.Sprintf("%s (%s)", k, strings.Join(v, ","))
			groups = append(groups, line)
		}
		row = append(row, strings.Join(users, "\n"))
		row = append(row, strings.Join(groups, "\n"))
		_ = output.AddRow(row)
		output.Draw()
		fmt.Println("Legend: m=admin; d=delete; w=deploy; n=annotate; r=read")
		os.Exit(0)
	}
}
