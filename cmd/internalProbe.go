package cmd

import (
	"strings"

	"github.com/robocorp/rcc/operations"
	"github.com/robocorp/rcc/pretty"
	"github.com/spf13/cobra"
)

func fixHosts(hosts []string) []string {
	servers := make([]string, len(hosts))
	for at, host := range hosts {
		if strings.Contains(host, ":") {
			servers[at] = host
		} else {
			servers[at] = host + ":443"
		}
	}
	return servers
}

var internalProbeCmd = &cobra.Command{
	Use:   "probe <host:port>+",
	Short: "Probe host:port combinations for supported TLS versions.",
	Long:  "Probe host:port combinations for supported TLS versions.",
	Run: func(cmd *cobra.Command, args []string) {
		servers := fixHosts(args)
		err := operations.TLSProbe(servers)
		pretty.Guard(err == nil, 1, "Probe failure: %v", err)
		pretty.Ok()
	},
}

func init() {
	internalCmd.AddCommand(internalProbeCmd)
}
