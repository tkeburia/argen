// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/gobuffalo/packr"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
)

type FileDescription struct {
	DestinationFileName string `yaml:"destinationFileName"`
	DestinationFilePath string `yaml:"destinationFilePath"`
	Template            string `yaml:"template"`
}

// moduleCmd represents the module command
var moduleCmd = &cobra.Command{
	Use:   "module module_name",
	Short: "Generate module",
	Long:  ``,
	Run:   genModule,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("requires a module name")
		}
		return nil
	},
}

var templates = packr.NewBox("../files/templates")
var staticFiles = packr.NewBox("../files/static")

func init() {
	rootCmd.AddCommand(moduleCmd)
}

func genModule(cmd *cobra.Command, args []string) {

	moduleNameCapitalCase := strings.Title(args[0])
	moduleNameLowerCase := strings.ToLower(args[0])

	templatesYaml, err := ioutil.ReadFile("/Users/tornikekeburia/sandbox/argen/files/templates.yml")
	check(err)

	var templateFileNames []FileDescription

	err = yaml.Unmarshal(templatesYaml, &templateFileNames)
	check(err)

	for _, el := range templateFileNames {
		writeTemplate(el, moduleNameLowerCase, moduleNameCapitalCase)
	}

	staticYaml, err := ioutil.ReadFile("/Users/tornikekeburia/sandbox/argen/files/static.yml")
	check(err)

	var staticFileNames []FileDescription

	err = yaml.Unmarshal(staticYaml, &staticFileNames)
	check(err)

	for _, el := range staticFileNames {
		writeStatic(el, moduleNameLowerCase, moduleNameCapitalCase)
	}

	fmt.Printf("Don't forget to add the following lines:\n\t"+
		"const val feature%s = \":feature_%s\" to ModuleDependency.kt\n\t"+
		"ModuleDependency.feature%s to settings.gradle.kts\n\t"+
		"import(%sFeatureModule) to BaseApplication.kt\n",
		moduleNameCapitalCase, moduleNameLowerCase, moduleNameCapitalCase, moduleNameLowerCase)
}

func writeTemplate(f FileDescription, moduleNameLowerCase string, moduleNameCapitalCase string) {
	var data bytes.Buffer
	argMap := map[string]interface{}{
		"ModuleNameLowerCase":   moduleNameLowerCase,
		"ModuleNameCapitalCase": moduleNameCapitalCase,
	}

	input, err := templates.FindString(f.Template)
	check(err)

	t := template.Must(template.New(f.Template).Parse(input))
	e := t.Execute(&data, argMap)
	check(e)

	resolvedPath := resolveString(f.DestinationFilePath, argMap)
	resolvedFileName := resolveString(f.DestinationFileName, argMap)

	check(os.MkdirAll(resolvedPath, os.ModePerm))

	e = ioutil.WriteFile(truncatingSprintf(resolvedPath+resolvedFileName, moduleNameCapitalCase), data.Bytes(), 0644)
	check(e)
}

func writeStatic(f FileDescription, moduleNameLowerCase string, moduleNameCapitalCase string) {
	argMap := map[string]interface{}{
		"ModuleNameLowerCase": moduleNameLowerCase,
		"ModuleNameCapitalCase": moduleNameCapitalCase,
	}
	resolvedPath := resolveString(f.DestinationFilePath, argMap)
	resolvedFileName := resolveString(f.DestinationFileName, argMap)

	check(os.MkdirAll(resolvedPath, os.ModePerm))

	input, err := staticFiles.Find(f.Template)
	check(err)

	e := ioutil.WriteFile(resolvedPath+resolvedFileName, input, 0644)
	check(e)
}

func resolveString(s string, argMap map[string]interface{}) string {
	ft := template.Must(template.New(s).Parse(s))
	var resolvedPath bytes.Buffer
	check(ft.Execute(&resolvedPath, argMap))
	return resolvedPath.String()
}

func truncatingSprintf(str string, args ...interface{}) string {
	n := strings.Count(str, "%s")
	if n > len(args) {
		panic("Unexpected string:" + str)
	}
	return fmt.Sprintf(str, args[:n]...)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
