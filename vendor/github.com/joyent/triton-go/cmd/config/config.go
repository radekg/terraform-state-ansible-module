package config

import (
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	triton "github.com/joyent/triton-go"
	"github.com/joyent/triton-go/authentication"
	"github.com/joyent/triton-go/cmd/internal/config"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type TritonClientConfig struct {
	Config *triton.ClientConfig
}

func buildSSHAgentSigner(keyID string, accountName string) (*authentication.SSHAgentSigner, error) {
	signer, err := authentication.NewSSHAgentSigner(authentication.SSHAgentSignerInput{
		KeyID:       keyID,
		AccountName: accountName,
	})
	if err != nil {
		return nil, err
	}
	return signer, nil
}

func buildPrivateKeySigner(keyID string, accountName string, keyMaterial []byte) (*authentication.PrivateKeySigner, error) {
	signer, err := authentication.NewPrivateKeySigner(authentication.PrivateKeySignerInput{
		KeyID:              keyID,
		PrivateKeyMaterial: keyMaterial,
		AccountName:        accountName,
	})
	if err != nil {
		return nil, err
	}
	return signer, nil
}

func getKeyMaterialContents(keyMaterial string) ([]byte, error) {
	var keyBytes []byte
	if _, err := os.Stat(keyMaterial); err == nil {
		keyBytes, err = ioutil.ReadFile(keyMaterial)
		if err != nil {
			return nil, fmt.Errorf("error reading key material from %s: %s",
				keyMaterial, err)
		}
		block, _ := pem.Decode(keyBytes)
		if block == nil {
			return nil, fmt.Errorf(
				"failed to read key material '%s': no key found", keyMaterial)
		}

		if block.Headers["Proc-Type"] == "4,ENCRYPTED" {
			return nil, fmt.Errorf(
				"failed to read key '%s': password protected keys are\n"+
					"not currently supported. Please decrypt the key prior to use.", keyMaterial)
		}

	} else {
		keyBytes = []byte(keyMaterial)
	}

	return keyBytes, nil
}

func NewTritonConfig() (*TritonClientConfig, error) {
	viper.AutomaticEnv()

	var signer authentication.Signer
	var err error

	keyMaterial := GetTritonKeyMaterial()
	if keyMaterial == "" {
		signer, err = buildSSHAgentSigner(GetTritonKeyID(), GetTritonAccount())
		if err != nil {
			log.Fatal().Str("func", "initConfig").Msg("Error Creating Triton SSH Agent Signer")
			return nil, err
		}
	} else {
		keyMaterial, err := getKeyMaterialContents(keyMaterial)
		if err != nil {
			return nil, err
		}

		signer, err = buildPrivateKeySigner(GetTritonKeyID(), GetTritonAccount(), keyMaterial)
		if err != nil {
			return nil, errors.Wrap(err, "Error Creating Triton SSH Private Key Signer")
		}
	}

	config := &triton.ClientConfig{
		TritonURL:   GetTritonURL(),
		AccountName: GetTritonAccount(),
		Signers:     []authentication.Signer{signer},
	}

	return &TritonClientConfig{
		Config: config,
	}, nil
}

func NewMantaConfig() (*TritonClientConfig, error) {
	viper.AutomaticEnv()

	var signer authentication.Signer
	var err error

	keyMaterial := GetMantaKeyMaterial()
	if keyMaterial == "" {
		signer, err = buildSSHAgentSigner(GetMantaKeyID(), GetMantaAccount())
		if err != nil {
			log.Fatal().Str("func", "initConfig").Msg("Error Creating Manta SSH Agent Signer")
			return nil, err
		}
	} else {
		keyMaterial, err := getKeyMaterialContents(keyMaterial)
		if err != nil {
			return nil, err
		}

		signer, err = buildPrivateKeySigner(GetMantaKeyID(), GetMantaAccount(), keyMaterial)
		if err != nil {
			return nil, errors.Wrap(err, "Error Creating Manta SSH Private Key Signer")
		}
	}

	config := &triton.ClientConfig{
		MantaURL:    GetMantaURL(),
		AccountName: GetMantaAccount(),
		Signers:     []authentication.Signer{signer},
	}

	return &TritonClientConfig{
		Config: config,
	}, nil
}

var tritonEnvPrefixes = []string{"TRITON", "SDC"}
var mantaEnvPrefixes = []string{"MANTA", "TRITON", "SDC"}

func getTritonEnvVar(name string) string {
	for _, prefix := range tritonEnvPrefixes {
		if val := viper.GetString(prefix + "_" + name); val != "" {
			return val
		}
	}

	return ""
}

func getMantaEnvVar(name string) string {
	for _, prefix := range mantaEnvPrefixes {
		if val := viper.GetString(prefix + "_" + name); val != "" {
			return val
		}
	}

	return ""
}

func GetTritonURL() string {
	url := viper.GetString(config.KeyTritonURL)
	if url == "" {
		url = getTritonEnvVar("URL")
	}

	return url
}

func GetTritonKeyMaterial() string {
	url := viper.GetString(config.KeyTritonSSHKeyMaterial)
	if url == "" {
		url = getTritonEnvVar("KEY_MATERIAL")
	}

	return url
}

func GetTritonAccount() string {
	account := viper.GetString(config.KeyTritonAccount)
	if account == "" {
		account = getTritonEnvVar("ACCOUNT")
	}

	return account
}

func GetTritonKeyID() string {
	keyID := viper.GetString(config.KeyTritonSSHKeyID)
	if keyID == "" {
		keyID = getTritonEnvVar("KEY_ID")
	}

	return keyID
}

func GetMantaURL() string {
	url := viper.GetString(config.KeyMantaURL)
	if url == "" {
		url = getMantaEnvVar("URL")
	}

	return url
}

func GetMantaKeyMaterial() string {
	url := viper.GetString(config.KeyMantaSSHKeyMaterial)
	if url == "" {
		url = getMantaEnvVar("KEY_MATERIAL")
	}

	return url
}

func GetMantaAccount() string {
	account := viper.GetString(config.KeyMantaAccount)
	if account == "" {
		account = getMantaEnvVar("USER")
	}

	return account
}

func GetMantaKeyID() string {
	keyID := viper.GetString(config.KeyMantaSSHKeyID)
	if keyID == "" {
		keyID = getMantaEnvVar("KEY_ID")
	}

	return keyID
}

func GetPkgID() string {
	return viper.GetString(config.KeyPackageID)
}

func GetPkgName() string {
	return viper.GetString(config.KeyPackageName)
}

func GetPkgMemory() int {
	return viper.GetInt(config.KeyPackageMemory)
}

func GetPkgDisk() int {
	return viper.GetInt(config.KeyPackageDisk)
}

func GetPkgSwap() int {
	return viper.GetInt(config.KeyPackageSwap)
}

func GetPkgVPCUs() int {
	return viper.GetInt(config.KeyPackageVPCUs)
}

func GetImgID() string {
	return viper.GetString(config.KeyImageId)
}

func GetImgName() string {
	return viper.GetString(config.KeyImageName)
}

func GetMachineID() string {
	return viper.GetString(config.KeyInstanceID)
}

func GetMachineName() string {
	return viper.GetString(config.KeyInstanceName)
}

func GetMachineState() string {
	return viper.GetString(config.KeyInstanceState)
}

func GetMachineBrand() string {
	return viper.GetString(config.KeyInstanceBrand)
}

func GetMachineFirewall() bool {
	return viper.GetBool(config.KeyInstanceFirewall)
}

func GetMachineNetworks() []string {
	if viper.IsSet(config.KeyInstanceNetwork) {
		var networks []string
		cfg := viper.GetStringSlice(config.KeyInstanceNetwork)
		for _, i := range cfg {
			networks = append(networks, i)
		}

		return networks
	}
	return nil
}

func GetMachineAffinityRules() []string {
	if viper.IsSet(config.KeyInstanceAffinityRule) {
		var rules []string
		cfg := viper.GetStringSlice(config.KeyInstanceAffinityRule)
		for _, i := range cfg {
			rules = append(rules, i)
		}

		return rules
	}
	return nil
}

func GetMachineTags() map[string]string {
	if viper.IsSet(config.KeyInstanceTag) {
		tags := make(map[string]string, 0)
		cfg := viper.GetStringSlice(config.KeyInstanceTag)
		for _, i := range cfg {
			m := strings.Split(i, "=")
			tags[m[0]] = m[1]
		}

		return tags
	}

	return nil
}

func GetSearchTags() map[string]interface{} {
	if viper.IsSet(config.KeyInstanceTag) {
		tags := make(map[string]interface{}, 0)
		cfg := viper.GetStringSlice(config.KeyInstanceTag)
		for _, i := range cfg {
			m := strings.Split(i, "=")
			tags[m[0]] = m[1]
		}

		return tags
	}

	return nil
}

func GetMachineMetadata() map[string]string {
	if viper.IsSet(config.KeyInstanceMetadata) {
		metadata := make(map[string]string, 0)
		cfg := viper.GetStringSlice(config.KeyInstanceMetadata)
		for _, i := range cfg {
			m := strings.Split(i, "=")
			metadata[m[0]] = m[1]
		}

		return metadata
	}

	return nil
}

func GetMachineUserdata() string {
	return viper.GetString(config.KeyInstanceUserdata)
}

func GetAccountEmail() string {
	return viper.GetString(config.KeyAccountEmail)
}

func GetAccountCompanyName() string {
	return viper.GetString(config.KeyAccountCompanyName)
}

func GetAccountFirstName() string {
	return viper.GetString(config.KeyAccountFirstName)
}

func GetAccountLastName() string {
	return viper.GetString(config.KeyAccountLastName)
}

func GetAccountAddress() string {
	return viper.GetString(config.KeyAccountAddress)
}

func GetAccountPostalCode() string {
	return viper.GetString(config.KeyAccountPostcode)
}

func GetAccountCity() string {
	return viper.GetString(config.KeyAccountCity)
}

func GetAccountState() string {
	return viper.GetString(config.KeyAccountState)
}

func GetAccountCountry() string {
	return viper.GetString(config.KeyAccountCountry)
}

func GetAccountPhone() string {
	return viper.GetString(config.KeyAccountPhone)
}

func GetAccountCNSEnabled() string {
	return viper.GetString(config.KeyAccountTritonCNSEnabled)
}

func GetSSHKeyName() string {
	return viper.GetString(config.KeySSHKeyName)
}
func GetSSHKeyFingerprint() string {
	return viper.GetString(config.KeySSHKeyFingerprint)
}

func GetSSHKey() string {
	return viper.GetString(config.KeySSHKey)
}

func IsBlockingAction() bool {
	return viper.GetBool(config.KeyInstanceWait)
}

func FormatTime(t time.Time) string {
	d := time.Since(t)

	timeSegs := make([]string, 0, 6)

	years := int64(float64(d/(24*time.Hour)) / 365.25)
	if years > 0 {
		timeSegs = append(timeSegs, fmt.Sprintf("%2dy", years))
	}

	months := int64(float64(d/(24*time.Hour)%365) / 30.25)
	if months > 0 {
		timeSegs = append(timeSegs, fmt.Sprintf("%2dmo", months))
	}

	days := int64(d/(24*time.Hour)) % 365 % 7
	if days > 0 {
		timeSegs = append(timeSegs, fmt.Sprintf("%2dd", days))
	}

	hours := int64(d.Hours()) % 24
	if hours > 0 {
		timeSegs = append(timeSegs, fmt.Sprintf("%2dh", hours))
	}

	minutes := int64(d.Minutes()) % 60
	if minutes > 0 {
		timeSegs = append(timeSegs, fmt.Sprintf("%2dm", minutes))
	}

	seconds := int64(d.Seconds()) % 60
	if seconds > 0 {
		timeSegs = append(timeSegs, fmt.Sprintf("%2ds", seconds))
	}

	maxSegs := len(timeSegs)
	if maxSegs > 1 {
		maxSegs = 1
	}

	timeSegs = timeSegs[0:maxSegs]

	return strings.Join(timeSegs, " ")
}
