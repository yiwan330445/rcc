package cmd

import (
	"os"
	"path/filepath"

	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/conda"
	"github.com/robocorp/rcc/htfs"
	"github.com/robocorp/rcc/journal"
	"github.com/robocorp/rcc/pathlib"
	"github.com/robocorp/rcc/pretty"
	"github.com/robocorp/rcc/shell"

	"github.com/spf13/cobra"
)

func deleteByExactIdentity(exact string) {
	_, roots := htfs.LoadCatalogs()
	for _, label := range roots.FindEnvironments([]string{exact}) {
		common.Log("Removing %v", label)
		err := roots.RemoveHolotreeSpace(label)
		pretty.Guard(err == nil, 4, "Error: %v", err)
	}
}

var holotreeVenvCmd = &cobra.Command{
	Use:   "venv conda.yaml+",
	Short: "Create user managed virtual python environment inside automation folder.",
	Long:  "Create user managed virtual python environment inside automation folder.",
	Args:  cobra.MinimumNArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		defer journal.BuildEventStats("venv")
		if common.DebugFlag() {
			defer common.Stopwatch("Holotree venv command lasted").Report()
		}

		// following settings are forced in venv environments
		common.UnmanagedSpace = true
		common.ExternallyManaged = true
		common.ControllerType = "venv"

		where, err := os.Getwd()
		pretty.Guard(err == nil, 1, "Error: %v", err)
		location := filepath.Join(where, "venv")

		previous := pathlib.IsDir(location)
		if holotreeForce && previous {
			pretty.Note("Trying to remove existing venv at %q ...", location)
			err := pathlib.TryRemoveAll("venv", location)
			pretty.Guard(err == nil, 2, "Error: %v", err)
		}

		pretty.Guard(!pathlib.Exists(location), 3, "Name %q aready exists! Remove it, or use force.", location)

		if holotreeForce {
			identity := htfs.ControllerSpaceName([]byte(common.ControllerIdentity()), []byte(common.HolotreeSpace))
			deleteByExactIdentity(identity)
		}

		env := holotreeExpandEnvironment(args, "", "", "", 0, holotreeForce)
		envPath := pathlib.EnvironmentPath(env)
		python, ok := envPath.Which("python", conda.FileExtensions)
		if !ok {
			python, ok = envPath.Which("python3", conda.FileExtensions)
		}
		pretty.Guard(ok, 5, "For some reason, could not find python executable in environment paths. Report a bug. PATH: %q", envPath)
		pretty.Note("Trying to make new venv at %q using %q ...", location, python)
		task := shell.New(env, ".", python, "-m", "venv", "--system-site-packages", location)
		code, err := task.Execute(false)
		pretty.Guard(err == nil, 6, "Error: %v", err)
		pretty.Guard(code == 0, 7, "Exit code %d from venv creation.", code)

		pretty.Highlight("New venv is located at %q. For activation, use venv/bin/activate scripts.", location)

		pretty.Ok()
	},
}

func init() {
	holotreeCmd.AddCommand(holotreeVenvCmd)
	rootCmd.AddCommand(holotreeVenvCmd)

	holotreeVenvCmd.Flags().StringVarP(&common.HolotreeSpace, "space", "s", "user", "Client specific name to identify this environment.")
	holotreeVenvCmd.Flags().BoolVarP(&holotreeForce, "force", "f", false, "Force environment creation by deleting unmanaged space. Dangerous, do not use unless you understand what it means.")
}
