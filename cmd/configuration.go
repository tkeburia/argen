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
	"errors"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/tkeburia/argen/log"
	"github.com/tkeburia/argen/util"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"os"
	"os/exec"
)

var configurationCmd = &cobra.Command{
	Use:   "configuration",
	Short: "Manage configurations",
	Long:  `Manage configurations`,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
		return
	},
}

var addSubCmd = &cobra.Command{
	Use:   "add name github_repo",
	Short: "Add configuration from github",
	Long: `
Add configuration from github

name - name of the configuration that can be used to generate code
github_repo - source repository that contains configuration files to be used for generation
`,
	Run: addConfig,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return errors.New(log.ErrorS("requires a configuration name and url"))
		}
		return nil
	},
}

var rmSubCmd = &cobra.Command{
	Use:   "rm config_name",
	Short: "Delete local configuration",
	Long:  `Delete local configuration`,
	Run:   rmConfig,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New(log.ErrorS("requires a configuration name"))
		}
		return nil
	},
}

var lsSubCmd = &cobra.Command{
	Use:   "ls",
	Short: "list local configurations",
	Long:  `list local configurations`,
	Run:   lsConfig,
}

var updateSubCmd = &cobra.Command{
	Use:   "update config_name",
	Short: "Pull the most recent version of given configuration from github",
	Long: `
Pull the most recent version of given configuration from github

IMPORTANT: this will overwrite all local changes to the configuration files
`,
	Run: updateConfig,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New(log.ErrorS("requires a configuration name"))
		}
		return nil
	},
}

var showSubCmd = &cobra.Command{
	Use:   "show config_name",
	Short: "show detailed information about a specific configuration",
	Long:  `show detailed information about a specific configuration`,
	Run:   showConfig,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New(log.ErrorS("requires a configuration name"))
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configurationCmd)
	configurationCmd.AddCommand(addSubCmd)
	configurationCmd.AddCommand(rmSubCmd)
	configurationCmd.AddCommand(lsSubCmd)
	configurationCmd.AddCommand(updateSubCmd)
	configurationCmd.AddCommand(showSubCmd)
}

func addConfig(_ *cobra.Command, args []string) {
	clone(args[1], util.ConfigurationPath(args[0]))
}

func rmConfig(_ *cobra.Command, args []string) {
	remove(util.ConfigurationPath(args[0]))
}

func lsConfig(_ *cobra.Command, args []string) {
	d, e := os.Open(util.BaseConfigPath())
	util.Check(e)

	names, e := d.Readdirnames(-1)
	util.Check(e)

	for _, el := range names {
		fmt.Println(el)
	}
}

func showConfig(_ *cobra.Command, args []string) {
	templateFiles := util.ReadFile(util.FullPath(args[0], TemplatesFileName))
	staticFiles := util.ReadFile(util.FullPath(args[0], StaticFileName))

	table := makeTable()

	var tEntries [][]string
	var sEntries [][]string

	populateEntries(templateFiles, &tEntries, staticFiles, &sEntries)

	table.AppendBulk(tEntries)
	table.AppendBulk(sEntries)
	remotes := getRemotes(args[0])

	logInfo(remotes, args[0])

	table.Render()
}

func updateConfig(_ *cobra.Command, args []string) {
	path := util.ConfigurationPath(args[0])

	g, e := git.PlainOpen(path)
	util.Check(e)

	e = g.Fetch(&git.FetchOptions{})
	if e == git.NoErrAlreadyUpToDate {
		log.Info("No new upstream changes\n")
	} else {
		util.Check(e)
	}

	w, e := g.Worktree()
	util.Check(e)

	rh, e := g.ResolveRevision(plumbing.Revision("origin/master"))
	util.Check(e)

	log.Info("resetting any local changes...\n")
	util.Check(w.Reset(&git.ResetOptions{Mode: git.HardReset, Commit: *rh}))
}

func populateEntries(templateFiles []util.FileDescription, tEntries *[][]string, staticFiles []util.FileDescription, sEntries *[][]string) {
	for _, el := range templateFiles {
		*tEntries = append(*tEntries, []string{el.Template, el.DestinationFilePath, el.DestinationFileName})
	}
	for _, el := range staticFiles {
		*sEntries = append(*sEntries, []string{el.Template, el.DestinationFilePath, el.DestinationFileName})
	}
}

func logInfo(remotes []*git.Remote, configName string) {
	log.Info("\nRemote(s):\n")
	for _, el := range remotes {
		log.Info("\t%s\n", el.Config().URLs)
	}
	log.Info("Location: %s\n", util.ConfigurationPath(configName))
	log.Info("Files:")
}

func getRemotes(configName string) []*git.Remote {
	repository, err := git.PlainOpen(util.ConfigurationPath(configName))
	util.Check(err)
	remotes, err := repository.Remotes()
	util.Check(err)
	return remotes
}

func makeTable() *tablewriter.Table {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Source", "Destination path", "Destination file name"})
	table.SetHeaderAlignment(tablewriter.ALIGN_CENTER)
	table.SetAutoWrapText(false)
	table.SetRowLine(true)
	table.SetAutoMergeCells(true)
	return table
}

func clone(name string, directory string) {
	if needsUpdate(directory) {
		remove(directory)
		log.Verbose("cloning %s to %s\n", name, directory)
		util.Check(exec.Command("git", "clone", name, directory).Run())
	}
}

func remove(directory string) {
	log.Verbose("deleting %s...\n", directory)
	util.Check(os.RemoveAll(directory))
}

func needsUpdate(path string) bool {
	log.Verbose("util.Check()ing for differences\n")
	revision := "master"
	remoteRevision := "origin/master"

	g, e := git.PlainOpen(path)
	if e != nil {
		return true
	}

	log.Verbose("git rev-parse %s", revision)

	h, e := g.ResolveRevision(plumbing.Revision(revision))
	util.Check(e)

	log.Verbose("%s\n", h.String())

	log.Verbose("git rev-parse %s", remoteRevision)

	rh, e := g.ResolveRevision(plumbing.Revision(remoteRevision))
	util.Check(e)

	log.Verbose("%s\n", rh.String())

	updateNeeded := rh.String() != h.String()

	log.Verbose("Based on above, will%supdate configuration...\n", boolMap[updateNeeded])

	return updateNeeded
}

var boolMap = map[bool]string{
	false: " not ",
	true:  " ",
}

