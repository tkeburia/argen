package util

import (
	"fmt"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type FileDescription struct {
	DestinationFileName string `yaml:"destinationFileName"`
	DestinationFilePath string `yaml:"destinationFilePath"`
	Template            string `yaml:"template"`
}

func ReadFile(path string) []FileDescription  {
	templatesYaml, err := ioutil.ReadFile(fmt.Sprintf(path))
	Check(err)

	var result []FileDescription

	err = yaml.Unmarshal(templatesYaml, &result)
	Check(err)

	return result
}

func FullPath(relPath string, file string) string {
	return fmt.Sprintf("%s/%s", ConfigurationPath(relPath) , file)
}

func ConfigurationPath(name string) string {
	return fmt.Sprintf("%s/%s", BaseConfigPath(), name)
}

func BaseConfigPath() string {
	return viper.GetString("configPath")
}

func Check(e error) {
	if e != nil {
		panic(e)
	}
}