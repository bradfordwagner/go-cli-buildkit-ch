//usr/bin/env go run "$0" "$@"; exit "$?"

/*
requires:
brew install pbcopy
*/

package main

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	setEnv bool // global flag to tell if we are using KUBECONFIG
	tmux   bool
)

var rootCmd = &cobra.Command{
	Use: "ks",
}

var localCmd = &cobra.Command{
	Use:   "local",
	Short: "sets a local file to be the kubecontext",
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{""}, cobra.ShellCompDirectiveFilterFileExt
	},
	Run: func(cmd *cobra.Command, args []string) {
		execute(true, args)
	},
}

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "display info about the current context",
	Run: func(cmd *cobra.Command, args []string) {
		kubeConfigPath := fmt.Sprintf("%s/config", extractKubeDir())
		if kubectx := os.Getenv("KUBECONFIG"); kubectx != "" {
			split := strings.Split(kubectx, "/")
			logrus.Infof("config_override=%s", split[len(split)-1])
		} else if readlink, err := os.Readlink(kubeConfigPath); err == nil {
			split := strings.Split(readlink, "/")
			logrus.Infof("config_file=%s", split[len(split)-1])
		}
	},
}

var kubeCmd = &cobra.Command{
	Use:   "kube",
	Short: "sets a file from ~/.kube to be kubecontext",
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		kubeDirPattern := extractKubeDir()

		filesInKubeDir := []string{}
		ignorePaths := map[string]bool{
			"cache/discovery": true,
			"cache/http":      true,
			"http-cache":      true,
			"switch-state":    true,
			"kubens":          true,
		}
		filepath.Walk(kubeDirPattern, func(path string, info fs.FileInfo, err error) error {
			var shouldIgnore bool
			for ignorePath, _ := range ignorePaths {
				if strings.Contains(path, ignorePath) {
					shouldIgnore = true
					break
				}
			}
			if !shouldIgnore {
				filesInKubeDir = append(filesInKubeDir, path)
			}
			return nil
		})

		return filesInKubeDir, cobra.ShellCompDirectiveDefault
	},
	Run: func(cmd *cobra.Command, args []string) {
		execute(false, args)
	},
}

func init() {
	rootCmd.AddCommand(localCmd, kubeCmd, infoCmd)
	rootCmd.PersistentFlags().BoolVarP(&setEnv, "setenv", "s", false, "copies export KUBECONTEXT")
	rootCmd.PersistentFlags().BoolVarP(&tmux, "tmux", "t", false, "executes export KUBECONTEXT in a new tmux pane")
}

func extractKubeDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logrus.WithError(err).Fatal()
	}
	return fmt.Sprintf("%s/.kube", homeDir)
}

func execute(isLocal bool, args []string) {
	maxLength := 1
	if tmux {
		maxLength = len(args)
	}

	for i := 0; i < maxLength; i++ {
		filePath := args[i]
		// local flag was passed in
		if isLocal {
			// find absolute path to provided relative file path
			if p, err := filepath.Abs(filePath); err != nil {
				logrus.WithError(err).Fatalf("could not find file=%s", filePath)
			} else {
				filePath = p
			}
		}
		logrus.Infof("context file=%s", filePath)

		if tmux {
			tmuxSplit(filePath)
		} else if setEnv {
			if err := execKubeContextCommand(filePath); err != nil {
				logrus.WithError(err).Fatal("could not copy export command")
			} else {
				logrus.Info("copied kubeconfig command to clipboard, paste when ready")
				os.Exit(0)
			}
		} else {
			// link .kube file with filePath
			kubeConfigPath := fmt.Sprintf("%s/config", extractKubeDir())
			_ = os.RemoveAll(kubeConfigPath) // we don't care if this file exists or not
			if err := os.Symlink(filePath, kubeConfigPath); err != nil {
				logrus.WithError(err).Fatal("could not link configurations")
			}
		}
	}
}

func tmuxSplit(path string) {
	setLayout := exec.Command("tmux", "select-layout", "tiled")
	setLayout.Run()

	splitw := exec.Command("tmux", "splitw", "-d", "-P", "-F", "#{pane_index}")
	output, err := splitw.Output()
	if err != nil {
		logrus.WithError(err).Error("couldn't run hello world")
	}
	paneIndex := strings.TrimSpace(string(output))

	tmuxSendToPane(paneIndex, fmt.Sprintf("export KUBECONFIG=%s", path))
}

func tmuxSendToPane(paneIndex, command string) {
	setKubeConfig := exec.Command("tmux", "send", "-t", paneIndex, command, "ENTER")
	_ = setKubeConfig.Start()
}

func execKubeContextCommand(filePath string) (err error) {
	copyText := fmt.Sprintf("export KUBECONFIG=%s", filePath)
	return clipboard.WriteAll(copyText)
}

func main() {
	// setup logrus
	logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.SetLevel(logrus.InfoLevel)

	// cobra
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
