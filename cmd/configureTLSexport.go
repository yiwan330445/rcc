package cmd

import (
	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/operations"
	"github.com/robocorp/rcc/pathlib"
	"github.com/robocorp/rcc/pretty"
	"github.com/spf13/cobra"
)

var (
	pemFile string
)

var tlsExportCmd = &cobra.Command{
	Use:   "tlsexport <configuration.yaml>+",
	Short: "Export TLS certificates from set of secure and unsecure URLs.",
	Long: `Export TLS certificates from set of secure and unsecure URLs.

CLI examples:
  rcc configuration tlsexport --pemfile export.pem robot_urls.yaml
  rcc configuration tlsexport --pemfile many.pem company_urls.yaml robot_urls.yaml more_urls.yaml


Configuration example in YAML format:
# trusted:
#     - https://api.eu1.robocorp.com/
#     - https://pypi.org/
#     - https://files.pythonhosted.org/
# untrusted:
#     - https://self-signed.badssl.com/
`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, configfiles []string) {
		pretty.Guard(!pathlib.IsFile(common.CaBundleFile()), 1, "Cannot create certificate bundle, while profile provides %q!", common.CaBundleFile())
		err := operations.TLSExport(pemFile, configfiles)
		pretty.Guard(err == nil, 2, "Probe failure: %v", err)
		pretty.Ok()
	},
}

func init() {
	configureCmd.AddCommand(tlsExportCmd)
	tlsExportCmd.Flags().StringVarP(&pemFile, "pemfile", "p", "", "Name of exported PEM file to write.")
	tlsExportCmd.MarkFlagRequired("pemfile")
}
