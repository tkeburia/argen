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
	"github.com/spf13/cobra"
	"github.com/tkeburia/argen/log"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"os"
	"os/exec"
)

var configurationCmd = &cobra.Command{
	Use:   "configuration",
	Short: "Manage configurations",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
		return
	},
}

var addSubCmd = &cobra.Command{
	Use:   "add name github_repo",
	Short: "Add configuration from github",
	Long: `
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
	Long:  ``,
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
	Long:  ``,
	Run:   lsConfig,
}

var updateSubCmd = &cobra.Command{
	Use:   "update config_name",
	Short: "Pull the most recent version of given configuration from github",
	Long: `
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
	Long: ` `,
	Run: showConfig,
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

func addConfig(cmd *cobra.Command, args []string) {
	clone(args[1], fmt.Sprintf("%s/%s", configPath(), args[0]))
}

func rmConfig(cmd *cobra.Command, args []string) {
	remove(fmt.Sprintf("%s/%s", configPath(), args[0]))
}

func lsConfig(cmd *cobra.Command, args []string) {
	d, e := os.Open(configPath())
	check(e)

	names, e := d.Readdirnames(-1)
	check(e)

	for _, el := range names {
		fmt.Println(el)
	}
}

func showConfig(cmd *cobra.Command, args []string) {

}

func updateConfig(cmd *cobra.Command, args []string) {
	path := fmt.Sprintf("%s/%s", configPath(), args[0])

	g, e := git.PlainOpen(path)
	check(e)

	e = g.Fetch(&git.FetchOptions{})
	if e == git.NoErrAlreadyUpToDate {
		log.Info("No new upstream changes\n")
	} else {
		check(e)
	}

	w, e := g.Worktree()
	check(e)

	rh, e := g.ResolveRevision(plumbing.Revision("origin/master"))
	check(e)

	log.Info("resetting any local changes...\n")
	check(w.Reset(&git.ResetOptions{Mode:git.HardReset, Commit: *rh}))
}

func clone(name string, directory string) {
	if needsUpdate(directory) {
		remove(directory)
		log.Verbose("cloning %s to %s\n", name, directory)
		check(exec.Command("git", "clone", name, directory).Run())
	}
}

func remove(directory string) {
	log.Verbose("deleting %s...\n", directory)
	check(os.RemoveAll(directory))
}

func needsUpdate(path string) bool {
	log.Verbose("Checking for differences\n")
	revision := "master"
	remoteRevision := "origin/master"

	g, e := git.PlainOpen(path)
	if e != nil {
		return true
	}

	log.Verbose("git rev-parse %s", revision)

	h, e := g.ResolveRevision(plumbing.Revision(revision))
	check(e)

	log.Verbose("%s\n", h.String())

	log.Verbose("git rev-parse %s", remoteRevision)

	rh, e := g.ResolveRevision(plumbing.Revision(remoteRevision))
	check(e)

	log.Verbose("%s\n", rh.String())

	updateNeeded := rh.String() != h.String()

	log.Verbose("Based on above, will%supdate configuration...\n", boolMap[updateNeeded])

	return updateNeeded
}

var boolMap = map[bool]string {
	false: " not ",
	true: " ",
}
