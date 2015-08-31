package plugin

import (
	pflag "github.com/ogier/pflag"
)

type Args struct {
	Network string
	Path    string
}

// ParseCommandLine parses the standard command line interface for a squaddie
// plugin
func ParseCommandLine(cmdLine []string) (Args, error) {
	var result Args
	flags := pflag.NewFlagSet("Plugin API", pflag.ContinueOnError)
	flags.StringVarP(&result.Network, "network", "n", "unix",
		"The network type")
	flags.StringVarP(&result.Path, "path", "p", "/var/sarge.sock",
		"The RPC endpoint")

	err := flags.Parse(cmdLine)
	return result, err
}
