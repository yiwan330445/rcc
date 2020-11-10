package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/conda"
	"github.com/robocorp/rcc/operations"
	"github.com/robocorp/rcc/pathlib"
	"github.com/robocorp/rcc/xviper"

	"github.com/spf13/cobra"
)

func toplevelCommands(parent *cobra.Command) {
	common.Log("\nToplevel commands")
	for _, child := range parent.Commands() {
		if child.Hidden || len(child.Commands()) > 0 {
			continue
		}
		name := strings.Split(child.Use, " ")
		common.Log("| %-14s  %s", name[0], child.Short)
	}
}

func commandTree(level int, prefix string, parent *cobra.Command) {
	if parent.Hidden {
		return
	}
	if level == 1 && len(parent.Commands()) == 0 {
		return
	}
	if level == 1 {
		common.Log("%s", strings.TrimSpace(prefix))
	}
	name := strings.Split(parent.Use, " ")
	label := fmt.Sprintf("%s%s", prefix, name[0])
	common.Log("%-16s  %s", label, parent.Short)
	indent := prefix + "| "
	for _, child := range parent.Commands() {
		commandTree(level+1, indent, child)
	}
}

var rootCmd = &cobra.Command{
	Use:   "rcc",
	Short: "rcc is environment manager for Robocorp Developer Tools",
	Long: `rcc provides support for creating and managing tasks,
communicating with Robocorp Cloud, and managing virtual environments where
tasks can be developed, debugged, and run.`,
	Run: func(cmd *cobra.Command, args []string) {
		commandTree(0, "", cmd.Root())
		toplevelCommands(cmd.Root())
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		common.Exit(1, "Error: [rcc %v] %v", common.Version, err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&controllerType, "controller", "user", "internal, DO NOT USE (unless you know what you are doing)")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $ROBOCORP/rcc.yaml)")

	rootCmd.PersistentFlags().BoolVarP(&common.Silent, "silent", "", false, "be less verbose on output")
	rootCmd.PersistentFlags().BoolVarP(&common.Separator, "separator", "", false, "write one line separator '--' before JSON output")
	rootCmd.PersistentFlags().BoolVarP(&pathlib.Lockless, "lockless", "", false, "do not use file locking ... DANGER!")
	rootCmd.PersistentFlags().BoolVarP(&common.NoCache, "nocache", "", false, "do not use cache for credentials and tokens, always request them from cloud")

	rootCmd.PersistentFlags().BoolVarP(&common.Debug, "debug", "", false, "to get debug output where available (not for production use)")
	rootCmd.PersistentFlags().BoolVarP(&common.Trace, "trace", "", false, "to get trace output where available (not for production use)")
}

func initConfig() {
	if cfgFile != "" {
		xviper.SetConfigFile(cfgFile)
	} else {
		xviper.SetConfigFile(filepath.Join(conda.RobocorpHome(), "rcc.yaml"))
	}

	common.UnifyVerbosityFlags()
	if len(controllerType) > 0 {
		operations.BackgroundMetric("rcc", "rcc.controlled.by", controllerType)
	}

	if common.Trace {
		common.Log("CLI command was: %#v", os.Args)
	}

	if common.Debug {
		common.Log("Using config file: %v", xviper.ConfigFileUsed())
	}
}
