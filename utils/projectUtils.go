package utils

import (
	"fmt"
	"os"
	"gopkg.in/yaml.v3"
	"path/filepath"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
)

type ProjectConfig struct {
	Description string `yaml:"description"`
	Name        string `yaml:"name"`
	Options     struct {
		Refresh string `yaml:"refresh"`
	} `yaml:"options"`
	Runtime  string `yaml:"runtime"`
	Template struct {
		Config struct {
			AWSRegion struct {
				Default     string `yaml:"default"`
				Description string `yaml:"description"`
				Secret      bool   `yaml:"secret"`
			} `yaml:"aws:region"`
			PulumiTags struct {
				AWSRegionDeployed string `yaml:"awsRegionDeployed"`
				ProjectName       string `yaml:"projectName"`
			} `yaml:"pulumi:tags"`
		} `yaml:"config"`
		Description string `yaml:"description"`
		DisplayName string `yaml:"displayName"`
		Metadata    struct {
			Cloud string `yaml:"cloud"`
		} `yaml:"metadata"`
	} `yaml:"template"`
}

func GeneratePulumiYAML(config ProjectConfig, filepath string) error {
	data, err := yaml.Marshal(&config)
	if err != nil {
		return fmt.Errorf("could not marshal YAML: %v", err)
	}
	err = os.WriteFile(filepath, data, 0644)
	if err != nil {
		return fmt.Errorf("could not write file: %v", err)
	}
	return nil
}

func BuildProjectConfig(projectName, awsRegion string) ProjectConfig {
	return ProjectConfig{
		Description: "Cloud project created by Maos Corp.",
		Name:        projectName,
		Runtime:     "go",
		Options: struct {
			Refresh string `yaml:"refresh"`
		}{
			Refresh: "always",
		},
		Template: struct {
			Config struct {
				AWSRegion struct {
					Default     string `yaml:"default"`
					Description string `yaml:"description"`
					Secret      bool   `yaml:"secret"`
				} `yaml:"aws:region"`
				PulumiTags struct {
					AWSRegionDeployed string `yaml:"awsRegionDeployed"`
					ProjectName       string `yaml:"projectName"`
				} `yaml:"pulumi:tags"`
			} `yaml:"config"`
			Description string `yaml:"description"`
			DisplayName string `yaml:"displayName"`
			Metadata    struct {
				Cloud string `yaml:"cloud"`
			} `yaml:"metadata"`
		}{
			Config: struct {
				AWSRegion struct {
					Default     string `yaml:"default"`
					Description string `yaml:"description"`
					Secret      bool   `yaml:"secret"`
				} `yaml:"aws:region"`
				PulumiTags struct {
					AWSRegionDeployed string `yaml:"awsRegionDeployed"`
					ProjectName       string `yaml:"projectName"`
				} `yaml:"pulumi:tags"`
			}{
				AWSRegion: struct {
					Default     string `yaml:"default"`
					Description string `yaml:"description"`
					Secret      bool   `yaml:"secret"`
				}{
					Default:     awsRegion,
					Description: "The AWS region to deploy to.",
					Secret:      true,
				},
				PulumiTags: struct {
					AWSRegionDeployed string `yaml:"awsRegionDeployed"`
					ProjectName       string `yaml:"projectName"`
				}{
					AWSRegionDeployed: awsRegion,
					ProjectName:       projectName,
				},
			},
			Description: "A brief description of the Environment name",
			DisplayName: "Environment Name, Prod,Staging",
			Metadata: struct {
				Cloud string `yaml:"cloud"`
			}{
				Cloud: "aws",
			},
		},
	}
}


// GetRootDir returns the root directory of the project
func GetRootDir() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	exeDir := filepath.Dir(exePath)
	return filepath.Join(exeDir, ".."), nil // Adjusting the path to the project directory
}

func ConvertProjectConfigToAutoConfigMap(projectConfig ProjectConfig) auto.ConfigMap {
	configMap := auto.ConfigMap{}

	configMap["description"] = auto.ConfigValue{Value: projectConfig.Description}
	configMap["name"] = auto.ConfigValue{Value: projectConfig.Name}
	configMap["runtime"] = auto.ConfigValue{Value: projectConfig.Runtime}
	configMap["options:refresh"] = auto.ConfigValue{Value: projectConfig.Options.Refresh}
	configMap["template:description"] = auto.ConfigValue{Value: projectConfig.Template.Description}
	configMap["template:displayName"] = auto.ConfigValue{Value: projectConfig.Template.DisplayName}
	configMap["template:metadata:cloud"] = auto.ConfigValue{Value: projectConfig.Template.Metadata.Cloud}
	configMap["template:config:aws:region:default"] = auto.ConfigValue{Value: projectConfig.Template.Config.AWSRegion.Default}
	configMap["template:config:aws:region:description"] = auto.ConfigValue{Value: projectConfig.Template.Config.AWSRegion.Description}
	configMap["template:config:aws:region:secret"] = auto.ConfigValue{Value: fmt.Sprintf("%v", projectConfig.Template.Config.AWSRegion.Secret)}
	configMap["template:config:pulumi:tags:awsRegionDeployed"] = auto.ConfigValue{Value: projectConfig.Template.Config.PulumiTags.AWSRegionDeployed}
	configMap["template:config:pulumi:tags:projectName"] = auto.ConfigValue{Value: projectConfig.Template.Config.PulumiTags.ProjectName}

	return configMap
}
