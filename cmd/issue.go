package cmd

import (
	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/operations"
	"github.com/robocorp/rcc/pretty"

	"github.com/spf13/cobra"
)

var (
	issueRobot       string
	issueMetafile    string
	issueAttachments []string
)

var issueCmd = &cobra.Command{
	Use:   "issue",
	Short: "Send an issue to Robocorp Control Room via rcc.",
	Long:  "Send an issue to Robocorp Control Room via rcc.",
	Run: func(cmd *cobra.Command, args []string) {
		if common.DebugFlag() {
			defer common.Stopwatch("Feedback issue lasted").Report()
		}
		accountEmail := "unknown"
		account := operations.AccountByName(AccountName())
		if account != nil && account.Details != nil {
			email, ok := account.Details["email"].(string)
			if ok {
				accountEmail = email
			}
		}
		err := operations.ReportIssue(accountEmail, issueRobot, issueMetafile, issueAttachments, dryFlag)
		if err != nil {
			pretty.Exit(1, "Error: %s", err)
		}
		pretty.Exit(0, "OK")
	},
}

func init() {
	feedbackCmd.AddCommand(issueCmd)
	issueCmd.Flags().StringVarP(&issueMetafile, "report", "r", "", "Report file in JSON form containing actual issue report details.")
	issueCmd.MarkFlagRequired("report")
	issueCmd.Flags().StringArrayVarP(&issueAttachments, "attachments", "a", []string{}, "Files to attach to issue report.")
	issueCmd.Flags().BoolVarP(&dryFlag, "dryrun", "d", false, "Don't send issue report, just show what would report be.")
	issueCmd.Flags().StringVarP(&issueRobot, "robot", "", "", "Full path to 'robot.yaml' configuration file. [optional]")
}
