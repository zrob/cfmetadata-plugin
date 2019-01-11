package main

import (
	"code.cloudfoundry.org/cli/plugin"
	"fmt"
)

type CFMetadataPlugin struct{}

func main() {
	plugin.Start(new(CFMetadataPlugin))
}

func (c *CFMetadataPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	fmt.Println("Hello world")
}

func (c *CFMetadataPlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "cfmetadata",
		Version: plugin.VersionType{
			Major: 0,
			Minor: 1,
			Build: 0,
		},
		Commands: []plugin.Command{
			{
				Name:     "annotations",
				HelpText: "view or modify annotations for an API resource",
				UsageDetails: plugin.Usage{
					Usage: "cf annotations TODO: add usage",
				},
			},
		},
	}
}
