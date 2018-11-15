package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	artifactory "github.com/lusis/go-artifactory/artifactory.v54"
	"github.com/lusis/outputter"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	format = kingpin.Flag("format", "format to display output").
		Default("table").
		Enum(outputter.GetOutputters()...)
	user = kingpin.Arg("user", "User name to show").Required().String()
)

func main() {
	kingpin.Parse()
	output, _ := outputter.NewOutputter(*format)
	client, clientErr := artifactory.NewClientFromEnv()
	if clientErr != nil {
		fmt.Printf("%s\n", clientErr.Error())
		os.Exit(1)
	}
	u, err := client.GetUserDetails(*user, make(map[string]string))
	if err != nil {
		fmt.Printf("%s\n", err)
		os.Exit(1)
	} else {

		output.SetHeaders([]string{
			"Name",
			"Email",
			"Password",
			"Admin?",
			"Updatable?",
			"Last Logged In",
			"Internal Password Disabled?",
			"Realm",
			"Groups",
		})
		_ = output.AddRow([]string{
			u.Name,
			u.Email,
			"<hidden>",
			strconv.FormatBool(u.Admin),
			strconv.FormatBool(u.ProfileUpdatable),
			u.LastLoggedIn,
			strconv.FormatBool(u.InternalPasswordDisabled),
			u.Realm,
			strings.Join(u.Groups, ","),
		})
		output.Draw()
		os.Exit(0)
	}
}
