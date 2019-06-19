package util

import (
	"fmt"
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
	check(err)

	var result []FileDescription

	err = yaml.Unmarshal(templatesYaml, &result)
	check(err)

	return result
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}