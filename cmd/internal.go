package cmd

import (
	"github.com/spf13/cobra"
)

var internalCmd = &cobra.Command{
	Use:     "internal",
	Aliases: []string{"i"},
	Short:   "Internal commands.",
	Long:    `Internal commands. Kind of lego blocks.`,
	Hidden:  true,
}

func init() {
	rootCmd.AddCommand(internalCmd)
	internalCmd.PersistentFlags().StringVarP(&wskey, "wskey", "", "", "Cloud API workspace key (authorization).")
}
