package cmd

import (
	"fmt"

	"github.com/robocorp/rcc/cloud"
	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/operations"
	"github.com/robocorp/rcc/pretty"

	"github.com/spf13/cobra"
)

var (
	encryptionVersion int
)

var e2eeCmd = &cobra.Command{
	Use:   "encryption",
	Short: "Internal end-to-end encryption tester method",
	Long:  "Internal end-to-end encryption tester method",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if common.DebugFlag() {
			defer common.Stopwatch("Encryption lasted").Report()
		}
		if encryptionVersion == 1 {
			version1encryption(args)
		} else {
			version2encryption(args)
		}
		pretty.Ok()
	},
}

func version1encryption(args []string) {
	account := operations.AccountByName(AccountName())
	pretty.Guard(account != nil, 1, "Could not find account by name: %q", AccountName())

	client, err := cloud.NewClient(account.Endpoint)
	pretty.Guard(err == nil, 2, "Could not create client for endpoint: %v, reason: %v", account.Endpoint, err)

	key, err := operations.GenerateEphemeralKey()
	pretty.Guard(err == nil, 3, "Problem with key generation, reason: %v", err)

	request := client.NewRequest("/assistant-v1/test/encryption")
	request.Body, err = key.RequestBody(args[0])
	pretty.Guard(err == nil, 4, "Problem with body generation, reason: %v", err)

	response := client.Post(request)
	pretty.Guard(response.Status == 200, 5, "Problem with test request, status=%d, body=%s", response.Status, response.Body)

	plaintext, err := key.Decode(response.Body)
	pretty.Guard(err == nil, 6, "Decode problem with body %s, reason: %v", response.Body, err)

	common.Log("Response: %s", string(plaintext))
}

func version2encryption(args []string) {
	account := operations.AccountByName(AccountName())
	pretty.Guard(account != nil, 1, "Could not find account by name: %q", AccountName())

	client, err := cloud.NewClient(account.Endpoint)
	pretty.Guard(err == nil, 2, "Could not create client for endpoint: %v, reason: %v", account.Endpoint, err)

	key, err := operations.GenerateEphemeralEccKey()
	pretty.Guard(err == nil, 3, "Problem with key generation, reason: %v", err)

	location := fmt.Sprintf("/assistant-v1/workspaces/%s/assistants/%s/test", workspaceId, assistantId)
	request := client.NewRequest(location)
	request.Headers["Authorization"] = fmt.Sprintf("RC-WSKEY %s", wskey)
	request.Body, err = key.RequestBody(nil)
	pretty.Guard(err == nil, 4, "Problem with body generation, reason: %v", err)

	common.Timeline("POST to cloud started")
	response := client.Post(request)
	common.Timeline("POST done")
	pretty.Guard(response.Status == 200, 5, "Problem with test request, status=%d, body=%s", response.Status, response.Body)

	common.Timeline("decode start")
	plaintext, err := key.Decode(response.Body)
	common.Timeline("decode done")
	pretty.Guard(err == nil, 6, "Decode problem with body %s, reason: %v", response.Body, err)

	common.Log("Response: %s", string(plaintext))
}

func init() {
	internalCmd.AddCommand(e2eeCmd)
	e2eeCmd.Flags().StringVarP(&accountName, "account", "a", "", "Account used for Robocorp Control Room operations.")
	e2eeCmd.Flags().IntVarP(&encryptionVersion, "use", "u", 1, "Which version of encryption method to test (1 or 2)")
	e2eeCmd.Flags().StringVarP(&workspaceId, "workspace", "", "", "Workspace id to get assistant information.")
	e2eeCmd.Flags().StringVarP(&assistantId, "assistant", "", "", "Assistant id to execute.")
}
