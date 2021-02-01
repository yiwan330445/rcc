package cmd

import (
	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/operations"
	"github.com/robocorp/rcc/pretty"

	"github.com/spf13/cobra"
)

var (
	issueMetafile    string
	issueAttachments []string
)

var issueCmd = &cobra.Command{
	Use:   "issue",
	Short: "Send an issue to Robocorp Cloud via rcc.",
	Long:  "Send an issue to Robocorp Cloud via rcc.",
	Run: func(cmd *cobra.Command, args []string) {
		if common.DebugFlag {
			defer common.Stopwatch("Feedback issue lasted").Report()
		}
		err := operations.ReportIssue(issueMetafile, issueAttachments, dryFlag)
		if err != nil {
			pretty.Exit(1, "Error: %s", err)
		}
		pretty.Exit(0, "OK")
	},
}

func init() {
	feedbackCmd.AddCommand(issueCmd)
	issueCmd.Flags().StringVarP(&issueMetafile, "report", "r", "", "Report file in JSON form containing actual issue report details.")
	issueCmd.Flags().StringArrayVarP(&issueAttachments, "attachments", "a", []string{}, "Files to attach to issue report.")
	issueCmd.Flags().BoolVarP(&dryFlag, "dryrun", "d", false, "Don't send issue report, just show what would report be.")
}
