package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/hashicorp/terraform/backend"

	tfBackendAtlas "github.com/hashicorp/terraform/backend/atlas"
	tfBackendLocal "github.com/hashicorp/terraform/backend/local"
	tfBackendRemote "github.com/hashicorp/terraform/backend/remote"

	tfBackendAzure "github.com/hashicorp/terraform/backend/remote-state/azure"
	tfBackendConsul "github.com/hashicorp/terraform/backend/remote-state/consul"
	tfBackendEtcdv3 "github.com/hashicorp/terraform/backend/remote-state/etcdv3"
	tfBackendGcs "github.com/hashicorp/terraform/backend/remote-state/gcs"
	tfBackendInmem "github.com/hashicorp/terraform/backend/remote-state/inmem"
	tfBackendManta "github.com/hashicorp/terraform/backend/remote-state/manta"
	tfBackendS3 "github.com/hashicorp/terraform/backend/remote-state/s3"
	tfBackendSwift "github.com/hashicorp/terraform/backend/remote-state/swift"

	tfConfig "github.com/hashicorp/terraform/config"
	tfState "github.com/hashicorp/terraform/state"
	"github.com/hashicorp/terraform/svchost/disco"
	"github.com/hashicorp/terraform/terraform"
)

// Retrieve represents an item to retrieve.
type Retrieve struct {
	ModulePath string `json:"module_path"`
	Retrieve   string `json:"retrieve"`
}

// ModuleArgs represents Ansible module arguments.
type ModuleArgs struct {
	TerraformConfigPath string     `json:"terraform_config_path"`
	State               string     `json:"state"`
	Retrieves           []Retrieve `json:"retrieves"`
	RequireAll          bool       `json:"require_all"`
}

// Response represents Ansible JSON message response.
type Response struct {
	Msg     string `json:"msg,omitempty"`
	Changed bool   `json:"changed"`
	Failed  bool   `json:"failed"`
}

type parsedResource struct {
	id         string
	attributes map[string]string
}

type parsedState struct {
	outputs   map[string]interface{}
	resources map[string]parsedResource
}

func exitJSON(responseBody Response) {
	returnResponse(responseBody)
}

func failJSON(responseBody Response) {
	responseBody.Failed = true
	returnResponse(responseBody)
}

func returnResponse(responseBody Response) {
	var response []byte
	var err error
	response, err = json.Marshal(responseBody)
	if err != nil {
		response, _ = json.Marshal(Response{Msg: "Invalid response object"})
	}
	fmt.Println(string(response))
	if responseBody.Failed {
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}

const (
	defaultState      = "default"
	defaultModulePath = "root"
)

func backendFromConfig(b backend.Backend, c map[string]interface{}) backend.Backend {
	rc, err := tfConfig.NewRawConfig(c)
	if err != nil {
		failJSON(Response{
			Msg: fmt.Sprintf("Error while configuring Terraform backend: '%+v'.", err)})
	}
	conf := terraform.NewResourceConfig(rc)
	_, errs := b.Validate(conf)
	if len(errs) > 0 {
		failJSON(Response{
			Msg: fmt.Sprintf("Error while configuring Terraform backend: '%+v'.", errs)})
	}
	if err := b.Configure(conf); err != nil {
		failJSON(Response{
			Msg: fmt.Sprintf("Error while configuring Terraform backend: '%+v'.", err)})
	}
	return b
}

func getBackend(bt string) backend.Backend {
	backends := map[string]interface{}{
		"azure":  func() backend.Backend { return tfBackendAzure.New() },
		"consul": func() backend.Backend { return tfBackendConsul.New() },
		"etcdv3": func() backend.Backend { return tfBackendEtcdv3.New() },
		"gcs":    func() backend.Backend { return tfBackendGcs.New() },
		"inmem":  func() backend.Backend { return tfBackendInmem.New() },
		"manta":  func() backend.Backend { return tfBackendManta.New() },
		"s3":     func() backend.Backend { return tfBackendS3.New() },
		"swift":  func() backend.Backend { return tfBackendSwift.New() },
	}
	b, ok := backends[bt]
	if !ok {
		failJSON(Response{
			Msg: fmt.Sprintf("Unknown remote-backend type '%s'.", bt)})
	}
	return b.(func() backend.Backend)()
}

func reduceToMap(bits []string, value interface{}, into map[string]interface{}) map[string]interface{} {
	if len(bits) == 1 {
		into[bits[0]] = value
		return into
	}
	if existing, ok := into[bits[0]]; ok {
		into[bits[0]] = reduceToMap(bits[1:], value, existing.(map[string]interface{}))
	} else {
		into[bits[0]] = reduceToMap(bits[1:], value, make(map[string]interface{}))
	}
	return into
}

func processState(state tfState.State, moduleArgs ModuleArgs) {
	ps := parsedState{
		outputs:   make(map[string]interface{}),
		resources: make(map[string]parsedResource),
	}

	state.RefreshState()

	for _, moduleState := range state.State().Modules {
		for key, outputState := range moduleState.Outputs {
			ps.outputs[fmt.Sprintf("%s.%s", strings.Join(moduleState.Path, "."), key)] = outputState.Value
		}
		for key, resourceState := range moduleState.Resources {
			ps.resources[fmt.Sprintf("%s.%s", strings.Join(moduleState.Path, "."), key)] = parsedResource{
				id:         resourceState.Primary.ID,
				attributes: resourceState.Primary.Attributes,
			}
		}
	}

	responseItems := make(map[string]interface{})

	// we have already validated these:
	for _, retrieve := range moduleArgs.Retrieves {
		segments := strings.Split(retrieve.Retrieve, "/")
		if segments[0] == "o" {
			lookupPath := fmt.Sprintf("%s.%s", retrieve.ModulePath, segments[1])
			responsePath := lookupPath
			if strings.HasPrefix(responsePath, fmt.Sprintf("%s.", defaultModulePath)) {
				responsePath = strings.Replace(responsePath, fmt.Sprintf("%s.", defaultModulePath), "", 1)
			}
			if v, ok := ps.outputs[lookupPath]; ok {
				responseItems[responsePath] = v
			} else {
				if moduleArgs.RequireAll {
					failJSON(Response{
						Msg: fmt.Sprintf("Output '%s' not found.", lookupPath)})
				}
			}
		}
		if segments[0] == "r" {
			lookupPath := fmt.Sprintf("%s.%s", retrieve.ModulePath, segments[1])
			responsePath := lookupPath
			if strings.HasPrefix(responsePath, fmt.Sprintf("%s.", defaultModulePath)) {
				responsePath = strings.Replace(responsePath, fmt.Sprintf("%s.", defaultModulePath), "", 1)
			}
			if v, ok := ps.resources[lookupPath]; ok {
				if a, ok2 := v.attributes[segments[2]]; ok2 {
					responseItems[fmt.Sprintf("%s.%s", responsePath, segments[2])] = a
				} else {
					if moduleArgs.RequireAll {
						failJSON(Response{
							Msg: fmt.Sprintf("Resource attribute '%s' not found.", segments[2])})
					}
				}
			} else {
				if moduleArgs.RequireAll {
					failJSON(Response{
						Msg: fmt.Sprintf("Resource '%s' not found.", lookupPath)})
				}
			}
		}
	}

	responseData := make(map[string]interface{})
	for k, v := range responseItems {
		responseData = reduceToMap(strings.Split(k, "."), v, responseData)
	}

	bytes, err := json.Marshal(responseData)
	if err != nil {
		failJSON(Response{
			Msg: fmt.Sprintf("Error while serializing response data. Reason: '%+v'.", err)})
	}
	exitJSON(Response{Msg: string(bytes)})
}

func validateRetrieve(retrieve string) {
	if strings.HasPrefix(retrieve, "o/") {
		if len(strings.Split(retrieve, "/")) != 2 {
			failJSON(Response{
				Msg: fmt.Sprintf("Output '%s' lookup format incorrect. Must be o/<name>.", retrieve)})
		}
	} else if strings.HasPrefix(retrieve, "r/") {
		if len(strings.Split(retrieve, "/")) != 3 {
			failJSON(Response{
				Msg: fmt.Sprintf("Resource '%s' lookup format incorrect. Must be r/<resource>/<property>.", retrieve)})
		}
	} else {
		failJSON(Response{
			Msg: fmt.Sprintf("Unsupported retrieve format: '%s'. Must start with 'o/' or 'r/'.", retrieve)})
	}
}

func handleOsArgs(osargs []string) ModuleArgs {

	if len(osargs) != 2 {
		failJSON(Response{
			Msg: "No argument file provided."})
	}
	argsFile := osargs[1]

	text, err := ioutil.ReadFile(argsFile)
	if err != nil {
		failJSON(Response{
			Msg: fmt.Sprintf("Could not read configuration file: '%s'. Reason: '%+v'.", argsFile, err)})
	}

	var args ModuleArgs
	err = json.Unmarshal(text, &args)
	if err != nil {
		failJSON(Response{
			Msg: fmt.Sprintf("Configuration file not valid JSON: '%s'. Reason: '%+v'.", argsFile, err)})
	}

	if args.TerraformConfigPath == "" {
		failJSON(Response{
			Msg: "Terraform configuration file not given. Missing terraform_config_path?"})
	}
	if len(args.Retrieves) == 0 {
		failJSON(Response{
			Msg: "Nothing to retrieve."})
	} else {
		newRetrieves := make([]Retrieve, 0)
		for _, r := range args.Retrieves {
			validateRetrieve(r.Retrieve)
			if r.ModulePath == "" {
				r.ModulePath = defaultModulePath
			}
			newRetrieves = append(newRetrieves, r)
		}
		args.Retrieves = newRetrieves
	}
	if args.State == "" {
		args.State = defaultState
	}

	return args
}

func executeProgram(osargs []string) {
	moduleArgs := handleOsArgs(osargs)

	notStandardBackends := map[string]bool{
		"atlas":  true,
		"remote": true,
		"local":  true,
	}

	cfg, err := tfConfig.LoadFile(moduleArgs.TerraformConfigPath)
	if err != nil {
		failJSON(Response{
			Msg: fmt.Sprintf("Terraform configuration file not found at: '%s'.", moduleArgs.TerraformConfigPath)})
	}

	if cfg.Terraform != nil {
		if cfg.Terraform.Backend != nil {

			errs := cfg.Terraform.Backend.Validate()
			if len(errs) > 0 {
				failJSON(Response{
					Msg: fmt.Sprintf("Error while validating the Terraform backend configuration: '%+v'.", errs)})
			}

			rawMap := cfg.Terraform.Backend.RawConfig.RawMap()

			if _, ok := notStandardBackends[cfg.Terraform.Backend.Type]; !ok {
				configuredBackend := backendFromConfig(getBackend(cfg.Terraform.Backend.Type), rawMap)
				state, err := configuredBackend.State(moduleArgs.State)
				if err != nil {
					failJSON(Response{
						Msg: fmt.Sprintf("Error while configuring Terraform backend: '%+v'.", err)})
				}
				processState(state, moduleArgs)

			} else {

				if cfg.Terraform.Backend.Type == "atlas" {
					atlasBacked := tfBackendAtlas.New()
					atlasRawConfig, err := tfConfig.NewRawConfig(rawMap)
					if err != nil {
						failJSON(Response{
							Msg: fmt.Sprintf("Error while configuring Terraform backend: '%+v'.", err)})
					}
					if err := atlasBacked.Configure(terraform.NewResourceConfig(atlasRawConfig)); err != nil {
						failJSON(Response{
							Msg: fmt.Sprintf("Error while configuring Terraform backend: '%+v'.", err)})
					}
					state, err := atlasBacked.State(moduleArgs.State)
					if err != nil {
						failJSON(Response{
							Msg: fmt.Sprintf("Error while configuring Terraform backend: '%+v'.", err)})
					}
					processState(state, moduleArgs)
				} else if cfg.Terraform.Backend.Type == "remote" {
					remoteBackend := tfBackendRemote.New(disco.New())
					remoteRawConfig, err := tfConfig.NewRawConfig(rawMap)
					if err != nil {
						failJSON(Response{
							Msg: fmt.Sprintf("Error while configuring Terraform backend: '%+v'.", err)})
					}
					if err := remoteBackend.Configure(terraform.NewResourceConfig(remoteRawConfig)); err != nil {
						failJSON(Response{
							Msg: fmt.Sprintf("Error while configuring Terraform backend: '%+v'.", err)})
					}
					state, err := remoteBackend.State(moduleArgs.State)
					if err != nil {
						failJSON(Response{
							Msg: fmt.Sprintf("Error while configuring Terraform backend: '%+v'.", err)})
					}
					processState(state, moduleArgs)
				} else if cfg.Terraform.Backend.Type == "local" {
					localBackend := tfBackendLocal.New()
					localRawConfig, err := tfConfig.NewRawConfig(rawMap)
					if err != nil {
						failJSON(Response{
							Msg: fmt.Sprintf("Error while configuring Terraform backend: '%+v'.", err)})
					}
					if err := localBackend.Configure(terraform.NewResourceConfig(localRawConfig)); err != nil {
						failJSON(Response{
							Msg: fmt.Sprintf("Error while configuring Terraform backend: '%+v'.", err)})
					}
					state, err := localBackend.State(moduleArgs.State)
					if err != nil {
						failJSON(Response{
							Msg: fmt.Sprintf("Error while configuring Terraform backend: '%+v'.", err)})
					}
					processState(state, moduleArgs)
				} else {
					failJSON(Response{
						Msg: fmt.Sprintf("Backend type '%s' not supported.", cfg.Terraform.Backend.Type)})
				}
			}
		}
	}
}

func main() {
	executeProgram(os.Args)
}
