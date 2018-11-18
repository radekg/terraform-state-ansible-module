package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
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
	"github.com/hashicorp/terraform/svchost/disco"
	"github.com/hashicorp/terraform/terraform"
)

// -- Ansible related:

// Retrieve represents an item to retrieve.
type Retrieve struct {
	ModulePath string `json:"module_path"`
	Retrieve   string `json:"retrieve"`
}

// ModuleArgs represents Ansible module arguments.
type ModuleArgs struct {
	TerraformFilePath string     `json:"terraform_file_path"`
	State             string     `json:"state"`
	Retrieves         []Retrieve `json:"retrieves"`
	RequireAll        bool       `json:"require_all"`
}

func (ma *ModuleArgs) handlingTerraformConfig() bool {
	return filepath.Base(ma.TerraformFilePath) == "terraform.tf"
}

func (ma *ModuleArgs) handlingTerraformVars() bool {
	return filepath.Base(ma.TerraformFilePath) == "vars.tf"
}

func (ma *ModuleArgs) validateRetrieve(retrieve string) error {
	if strings.HasPrefix(retrieve, "o/") {
		if len(strings.Split(retrieve, "/")) != 2 {
			return fmt.Errorf("Output '%s' lookup format incorrect. Must be o/<name>", retrieve)
		}
	} else if strings.HasPrefix(retrieve, "r/") {
		if len(strings.Split(retrieve, "/")) != 3 {
			return fmt.Errorf("Resource '%s' lookup format incorrect. Must be r/<resource>/<property>", retrieve)
		}
	} else {
		return fmt.Errorf("Unsupported retrieve format: '%s'. Must start with 'o/' or 'r/'", retrieve)
	}
	return nil
}

func (ma *ModuleArgs) validate() (*tfConfig.Config, error) {
	if !ma.handlingTerraformConfig() && !ma.handlingTerraformVars() {
		return nil, fmt.Errorf("Module supports terraform.tf and vars.tf files only")
	}
	if ma.TerraformFilePath == "" {
		return nil, fmt.Errorf("Terraform file path missing. terraform_file_path not set?")
	}
	cfg, err := tfConfig.LoadFile(ma.TerraformFilePath)
	if err != nil {
		return nil, fmt.Errorf("Terraform configuration file not found at: '%s'", ma.TerraformFilePath)
	}
	if ma.handlingTerraformConfig() {
		if len(ma.Retrieves) == 0 {
			return nil, fmt.Errorf("Nothing to retrieve")
		}
		for _, r := range ma.Retrieves {
			err := ma.validateRetrieve(r.Retrieve)
			if err != nil {
				return nil, err
			}
		}
	}
	return cfg, nil
}

// Response represents Ansible JSON message response.
type Response struct {
	Msg     string `json:"msg,omitempty"`
	Changed bool   `json:"changed"`
	Failed  bool   `json:"failed"`
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

// -- Module related:

type parsedResource struct {
	id         string
	attributes map[string]string
}

type parsedState struct {
	outputs   map[string]interface{}
	resources map[string]parsedResource
}

const (
	defaultState      = "default"
	defaultModulePath = "root"
)

var (
	irregularBackends = map[string]bool{
		"atlas": true,
	}
)

func maybeFailWithError(err error) {
	if err != nil {
		failJSON(Response{Msg: fmt.Sprintf("%+v", err)})
	}
}

func configureBackend(b backend.Backend, c map[string]interface{}) (backend.Backend, error) {
	rc, err := tfConfig.NewRawConfig(c)
	if err != nil {
		return backend.Nil{}, err
	}
	conf := terraform.NewResourceConfig(rc)
	_, errs := b.Validate(conf)
	if len(errs) > 0 {
		return backend.Nil{}, fmt.Errorf("Error while configuring Terraform backend: '%+v'", errs)
	}
	err = b.Configure(conf)
	if err != nil {
		return backend.Nil{}, err
	}
	return b, nil
}

func getBackend(bt string) (backend.Backend, error) {
	backends := map[string]interface{}{
		"azure":  func() backend.Backend { return tfBackendAzure.New() },
		"consul": func() backend.Backend { return tfBackendConsul.New() },
		"etcdv3": func() backend.Backend { return tfBackendEtcdv3.New() },
		"gcs":    func() backend.Backend { return tfBackendGcs.New() },
		"inmem":  func() backend.Backend { return tfBackendInmem.New() },
		"local":  func() backend.Backend { return tfBackendLocal.New() },
		"manta":  func() backend.Backend { return tfBackendManta.New() },
		"remote": func() backend.Backend { return tfBackendRemote.New(disco.New()) },
		"s3":     func() backend.Backend { return tfBackendS3.New() },
		"swift":  func() backend.Backend { return tfBackendSwift.New() },
	}
	b, ok := backends[bt]
	if !ok {
		return backend.Nil{}, fmt.Errorf("Unknown backend type '%s'", bt)
	}
	return b.(func() backend.Backend)(), nil
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

func processState(state *terraform.State, moduleArgs *ModuleArgs) (map[string]interface{}, error) {

	responseData := make(map[string]interface{})

	ps := parsedState{
		outputs:   make(map[string]interface{}),
		resources: make(map[string]parsedResource),
	}

	for _, moduleState := range state.Modules {
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
					return responseData, fmt.Errorf("Output '%s' not found", lookupPath)
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
						return responseData, fmt.Errorf("Resource attribute '%s' not found", segments[2])
					}
				}
			} else {
				if moduleArgs.RequireAll {
					return responseData, fmt.Errorf("Resource '%s' not found", lookupPath)
				}
			}
		}
	}

	for k, v := range responseItems {
		responseData = reduceToMap(strings.Split(k, "."), v, responseData)
	}

	return responseData, nil
}

func processVariables(vars []*tfConfig.Variable) (map[string]interface{}, error) {
	varsResponse := make(map[string]interface{})
	for _, v := range vars {
		switch v.Type() {
		case tfConfig.VariableTypeString:
			varsResponse[v.Name] = map[string]interface{}{
				"default": v.Default,
			}
		case tfConfig.VariableTypeMap:
			varsResponse[v.Name] = map[string]interface{}{
				"default": v.Default,
			}
		case tfConfig.VariableTypeList:
			varsResponse[v.Name] = map[string]interface{}{
				"default": v.Default,
			}
		default:
			return varsResponse, fmt.Errorf("Unsupported Terraform variable type '%s' for variable '%s'", v.DeclaredType, v.Name)
		}
	}
	return varsResponse, nil
}

func handleOsArgs(osargs []string) (*ModuleArgs, error) {
	if len(osargs) != 2 {
		return nil, fmt.Errorf("No argument file provided")
	}
	argsFile := osargs[1]
	text, err := ioutil.ReadFile(argsFile)
	if err != nil {
		return nil, fmt.Errorf("Could not read configuration file: '%s'. Reason: '%+v'", argsFile, err)
	}
	var args ModuleArgs
	err = json.Unmarshal(text, &args)
	if err != nil {
		return nil, fmt.Errorf("Configuration file not valid JSON: '%s'. Reason: '%+v'", argsFile, err)
	}
	newRetrieves := make([]Retrieve, 0)
	for _, r := range args.Retrieves {
		if r.ModulePath == "" {
			r.ModulePath = defaultModulePath
		}
		newRetrieves = append(newRetrieves, r)
	}
	args.Retrieves = newRetrieves
	if args.State == "" {
		args.State = defaultState
	}
	return &args, nil
}

func attemptFinishWithResponseData(responseData map[string]interface{}) ([]byte, error) {
	var bytes []byte
	bytes, err := json.Marshal(responseData)
	if err != nil {
		return bytes, fmt.Errorf("Error while serializing response data. Reason: '%+v'", err)
	}
	return bytes, nil
}

func executeProgram(osargs []string) {
	moduleArgs, err := handleOsArgs(osargs)
	maybeFailWithError(err)

	cfg, err := moduleArgs.validate()
	maybeFailWithError(err)

	if moduleArgs.handlingTerraformConfig() {
		if cfg.Terraform != nil {
			if cfg.Terraform.Backend != nil {

				errs := cfg.Terraform.Backend.Validate()
				if len(errs) > 0 {
					maybeFailWithError(fmt.Errorf("Error while validating the Terraform backend configuration: '%+v'", errs))
				}

				rawMap := cfg.Terraform.Backend.RawConfig.RawMap()

				if _, ok := irregularBackends[cfg.Terraform.Backend.Type]; !ok {
					b, err := getBackend(cfg.Terraform.Backend.Type)
					maybeFailWithError(err)
					configuredBackend, err := configureBackend(b, rawMap)
					maybeFailWithError(err)
					state, err := configuredBackend.State(moduleArgs.State)
					maybeFailWithError(err)
					state.RefreshState()
					responseData, err := processState(state.State(), moduleArgs)
					if err != nil {
						maybeFailWithError(err)
					}
					bytes, err := attemptFinishWithResponseData(responseData)
					if err != nil {
						maybeFailWithError(err)
					}
					exitJSON(Response{Msg: string(bytes)})
				} else {
					if cfg.Terraform.Backend.Type == "atlas" {
						atlasBacked := tfBackendAtlas.New()
						atlasRawConfig, err := tfConfig.NewRawConfig(rawMap)
						maybeFailWithError(err)
						err = atlasBacked.Configure(terraform.NewResourceConfig(atlasRawConfig))
						maybeFailWithError(err)
						state, err := atlasBacked.State(moduleArgs.State)
						maybeFailWithError(err)
						state.RefreshState()
						responseData, err := processState(state.State(), moduleArgs)
						if err != nil {
							maybeFailWithError(err)
						}
						bytes, err := attemptFinishWithResponseData(responseData)
						if err != nil {
							maybeFailWithError(err)
						}
						exitJSON(Response{Msg: string(bytes)})
					} else {
						maybeFailWithError(fmt.Errorf("Backend type '%s' not supported", cfg.Terraform.Backend.Type))
					}
				}
			}
		}
	} else if moduleArgs.handlingTerraformVars() {
		responseData, err := processVariables(cfg.Variables)
		if err != nil {
			maybeFailWithError(err)
		}
		bytes, err := attemptFinishWithResponseData(responseData)
		if err != nil {
			maybeFailWithError(err)
		}
		exitJSON(Response{Msg: string(bytes)})
	} else {
		maybeFailWithError(fmt.Errorf("Module supports terraform.tf and vars.tf files only"))
	}
}

func main() {
	executeProgram(os.Args)
}
