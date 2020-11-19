package cetokjob

import (
	"fmt"
	"io/ioutil"
	"regexp"

	"gopkg.in/yaml.v2"
)

// JobConfig represents the configuration of a job.
type JobConfig struct {
	Name      string `json:"name"`
	Image     string `json:"image"`
	Command   string `json:"command"`
	Configmap string `json:"configmap"`
	Secret    string `json:"secret"`
}

// LoadJobConfig reads the configuration file and returns []JobConfig
func LoadJobConfig(path string) ([]JobConfig, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	jobConfigs, err := readFile(buf)
	if err != nil {
		return nil, err
	}

	for _, jc := range jobConfigs {
		if err := validateJobConfig(jc); err != nil {
			return nil, fmt.Errorf("Config validate error: %s", err)
		}
	}

	return jobConfigs, nil
}

func readFile(fileBuffer []byte) ([]JobConfig, error) {
	data := []JobConfig{}
	err := yaml.Unmarshal(fileBuffer, &data)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal the config file : %s", err)
	}
	return data, nil
}

var dns1123SubdomainFmt = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`)

func validateJobConfig(jc JobConfig) error {
	if len(jc.Name) == 0 {
		return fmt.Errorf("name is required")
	}

	if len(jc.Name) > 43 {
		return fmt.Errorf("name must be no more than 43 characters")
	}

	if !dns1123SubdomainFmt.MatchString(jc.Name) {
		return fmt.Errorf("\"%s\" : DNS-1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character", jc.Name)
	}

	if len(jc.Image) == 0 {
		return fmt.Errorf("image is required")
	}
	return nil
}
