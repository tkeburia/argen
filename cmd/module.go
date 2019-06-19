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
	"github.com/spf13/cobra"
	"github.com/tkeburia/argen/util"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
)

var Configuration string

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

func init() {
	rootCmd.AddCommand(moduleCmd)
	moduleCmd.Flags().StringVarP(&Configuration, "configuration", "c", "", "Specify configuration to use")
}

func genModule(cmd *cobra.Command, args []string) {
	if Configuration == "" {
		check(cmd.Help())
		return
	}

	moduleNameCapitalCase := strings.Title(args[0])
	moduleNameLowerCase := strings.ToLower(args[0])

	var templateFileNames = util.ReadFile(fullPath(Configuration, "templates.yml"))

	for _, el := range templateFileNames {
		writeTemplate(el, moduleNameLowerCase, moduleNameCapitalCase)
	}

	var staticFileNames = util.ReadFile(fullPath(Configuration, "static.yml"))

	for _, el := range staticFileNames {
		writeStatic(el, moduleNameLowerCase, moduleNameCapitalCase)
	}

	fmt.Printf("Don't forget to add the following lines:\n\t"+
		"const val feature%s = \":feature_%s\" to ModuleDependency.kt\n\t"+
		"ModuleDependency.feature%s to settings.gradle.kts\n\t"+
		"import(%sFeatureModule) to BaseApplication.kt\n",
		moduleNameCapitalCase, moduleNameLowerCase, moduleNameCapitalCase, moduleNameLowerCase)
}

func writeTemplate(f util.FileDescription, moduleNameLowerCase string, moduleNameCapitalCase string) {
	var data bytes.Buffer
	argMap := map[string]interface{}{
		"ModuleNameLowerCase":   moduleNameLowerCase,
		"ModuleNameCapitalCase": moduleNameCapitalCase,
	}

	input, err := ioutil.ReadFile(fmt.Sprintf(fullPath(Configuration, f.Template)))
	check(err)

	t := template.Must(template.New(f.Template).Parse(string(input)))
	e := t.Execute(&data, argMap)
	check(e)

	resolvedPath := resolveString(f.DestinationFilePath, argMap)
	resolvedFileName := resolveString(f.DestinationFileName, argMap)

	check(os.MkdirAll(resolvedPath, os.ModePerm))

	e = ioutil.WriteFile(truncatingSprintf(resolvedPath+resolvedFileName, moduleNameCapitalCase), data.Bytes(), 0644)
	check(e)
}

func writeStatic(f util.FileDescription, moduleNameLowerCase string, moduleNameCapitalCase string) {
	argMap := map[string]interface{}{
		"ModuleNameLowerCase": moduleNameLowerCase,
		"ModuleNameCapitalCase": moduleNameCapitalCase,
	}
	resolvedPath := resolveString(f.DestinationFilePath, argMap)
	resolvedFileName := resolveString(f.DestinationFileName, argMap)

	check(os.MkdirAll(resolvedPath, os.ModePerm))

	input, err := ioutil.ReadFile(fmt.Sprintf(fullPath(Configuration, f.Template)))
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

func fullPath(relPath string, file string) string {
	return fmt.Sprintf("%s/%s/%s", configPath(), relPath, file)
}
