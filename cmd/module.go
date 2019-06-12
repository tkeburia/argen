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
	"io"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
)

type FileDescription struct {
	DestinationFileName string
	DestinationFilePath string
	SourceTemplateFile  string
	SourceTemplateName  string
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

func init() {
	rootCmd.AddCommand(moduleCmd)
}

func genModule(cmd *cobra.Command, args []string) {

	moduleNameCapitalCase := strings.Title(args[0])
	moduleNameLowerCase := strings.ToLower(args[0])

	for _, el := range templateFileNames {
		writeTemplate(el, moduleNameLowerCase, moduleNameCapitalCase)
	}

	for _, el := range staticFileNames {
		writeStatic(el, moduleNameLowerCase)
	}

	fmt.Printf("Don't forget to add the following lines: \n\tconst val "+
		"feature%s = \":feature_%s\" to ModuleDependency.kt \n\tModuleDependency.feature%s "+
		"to settings.gradle.kts\n", moduleNameCapitalCase, moduleNameLowerCase, moduleNameCapitalCase)
}

var templateFileNames = []FileDescription{
	{
		"AndroidManifest.xml",
		"feature_{{ .ModuleNameLowerCase }}/src/main/",
		"./cmd/templates/AndroidManifest",
		"AndroidManifest",
	},
	{
		"%sFeatureNavigator.kt",
		"feature_{{ .ModuleNameLowerCase }}/src/main/java/com/pagofx/feature/{{ .ModuleNameLowerCase }}/",
		"./cmd/templates/FeatureNavigator",
		"FeatureNavigator",
	},
	{
		"%sModule.kt",
		"feature_{{ .ModuleNameLowerCase }}/src/main/java/com/pagofx/feature/{{ .ModuleNameLowerCase }}/presentation/",
		"./cmd/templates/Module",
		"Module",
	},
}

var staticFileNames = []FileDescription{
	{
		"build.gradle.kts",
		"feature_{{ .ModuleNameLowerCase }}/",
		"./cmd/static/build.gradleFile",
		"",
	},
	{
		"lint.xml",
		"feature_{{ .ModuleNameLowerCase }}/",
		"./cmd/static/lint",
		"",
	},
	{
		".gitignore",
		"feature_{{ .ModuleNameLowerCase }}/",
		"./cmd/static/gitignore",
		"",
	},
}

func writeTemplate(f FileDescription, moduleNameLowerCase string, moduleNameCapitalCase string) {
	var data bytes.Buffer
	argMap := map[string]interface{}{
		"ModuleNameLowerCase":   moduleNameLowerCase,
		"ModuleNameCapitalCase": moduleNameCapitalCase,
	}

	t := template.Must(template.New(f.SourceTemplateName).ParseFiles(f.SourceTemplateFile))
	e := t.Execute(&data, argMap)
	check(e)

	ft := template.Must(template.New("path").Parse(f.DestinationFilePath))
	var resolvedPath bytes.Buffer
	check(ft.Execute(&resolvedPath, argMap))

	check(os.MkdirAll(resolvedPath.String(), os.ModePerm))

	e = ioutil.WriteFile(truncatingSprintf(resolvedPath.String()+f.DestinationFileName, moduleNameCapitalCase), data.Bytes(), 0644)
	check(e)
}

func writeStatic(f FileDescription, moduleNameLowerCase string) {
	argMap := map[string]interface{}{
		"ModuleNameLowerCase": moduleNameLowerCase,
	}

	ft := template.Must(template.New("path").Parse(f.DestinationFilePath))
	var resolvedPath bytes.Buffer
	check(ft.Execute(&resolvedPath, argMap))

	check(os.MkdirAll(resolvedPath.String(), os.ModePerm))

	in, err := os.Open(f.SourceTemplateFile)
	check(err)
	defer in.Close()

	out, err := os.Create(resolvedPath.String() + f.DestinationFileName)
	check(err)
	defer out.Close()

	_, e := io.Copy(out, in)
	check(e)
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
