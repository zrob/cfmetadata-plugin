package main

import (
	"code.cloudfoundry.org/cli/plugin"
	"encoding/json"
	"fmt"
	. "github.com/zrob/cfmetadata-plugin/util"
)

type CFMetadataPlugin struct{}

type ResourceList struct {
	Resources []ResourceModel `json:"resources"`
}

type ResourceModel struct {
	Guid     string        `json:"guid"`
	Metadata MetadataModel `json:"metadata"`
}

type MetadataModel struct {
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

func main() {
	plugin.Start(new(CFMetadataPlugin))
}

func (c *CFMetadataPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	argCount := len(args)

	if args[0] == "annotations" {
		if argCount < 3 || argCount > 3 {
			fmt.Println(c.GetMetadata().Commands[0].UsageDetails.Usage)
		} else {
			c.getAnnotations(cliConnection, args[1:])
		}
	}
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
					Usage: "cf annotations RESOURCE RESOURCE_NAME",
				},
			},
		},
	}
}

func (c *CFMetadataPlugin) getAnnotations(cliConnection plugin.CliConnection, args []string) {
	resource := args[0]
	name := args[1]

	output, err := cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("v3/%ss?names=%s", resource, name))
	FreakOut(err)

	response := parseCurlResponse(output)
	resources := ResourceList{}
	err = json.Unmarshal([]byte(response), &resources)
	FreakOut(err)

	if len(resources.Resources) == 0 {
		fmt.Printf("%s %s not found\r\n", resource, name)
		return
	} else if len(resources.Resources) > 1 {
		fmt.Printf("%s %s is ambiguous, more than one result returned\r\n", resource, name)
 		return
	} 

	fmt.Printf("Annotations for %s %s\r\n\r\n", resource, name)
	if len(resources.Resources[0].Metadata.Annotations) == 0 {
		fmt.Println("None")
	} else {
		for key, val := range resources.Resources[0].Metadata.Annotations {
			fmt.Printf("%s: %s\r\n", key, val)
		}
	}
}

func parseCurlResponse(output []string) string {
	var responseString string
	for _, part := range output {
		responseString += part
	}
	return responseString
}
