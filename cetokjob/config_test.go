package cetokjob

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestLoadJobConfig(t *testing.T) {
	t.Parallel()

	data := []struct {
		path       string
		jobConfigs []JobConfig
		err        error
	}{
		{
			path: "testdata/success.yaml",
			jobConfigs: []JobConfig{
				{
					Name:  "full-config",
					Image: "debian",
					Command: []string{
						"echo",
						"CONFIGMAP_SAMPLE:$(CONFIGMAP_SAMPLE) SECRET_SAMPLE:$(SECRET_SAMPLE) msg:$(msg)",
					},
					Args: []string{
						"args1",
						"args2",
					},
					Configmap: "sample-configmap",
					Secret:    "sample-secret",
				},
				{
					Name:  "simple-config",
					Image: "debian",
				},
				{
					// Test the maximum number of characters in Name
					Name:  "1000000000200000000030000000004000000000333",
					Image: "debian",
				},
			},
		},
		{
			path: "testdata/error-dns1123.yaml",
			err:  errors.New("Config validate error: \"Invalid-hostname\" : DNS-1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character"),
		},
		{
			path: "testdata/error-parse.yaml",
			err:  errors.New("Failed to unmarshal the config file : yaml: line 2: did not find expected key"),
		},
		{
			path: "testdata/error-no-name.yaml",
			err:  errors.New("Config validate error: name is required"),
		},
		{
			path: "testdata/error-name-over.yaml",
			err:  errors.New("Config validate error: name must be no more than 43 characters"),
		},
		{
			path: "testdata/error-image.yaml",
			err:  errors.New("Config validate error: image is required"),
		},
		{
			path: "testdata/not-found.yaml",
			err:  errors.New("open testdata/not-found.yaml: no such file or directory"),
		},
	}

	for _, single := range data {
		t.Run("", func(single struct {
			path       string
			jobConfigs []JobConfig
			err        error
		}) func(t *testing.T) {
			return func(t *testing.T) {
				jobConfigs, err := LoadJobConfig(single.path)
				if err != nil {
					if single.err == nil {
						t.Fatalf(err.Error())
					}
					if !strings.EqualFold(single.err.Error(), err.Error()) {
						t.Fatalf("expected err: %s got err: %s", single.err, err)
					}
				} else {
					if !reflect.DeepEqual(single.jobConfigs, jobConfigs) {
						t.Fatalf("expected %s pods, got %s", single.jobConfigs, jobConfigs)
					}
				}
			}
		}(single))
	}
}
