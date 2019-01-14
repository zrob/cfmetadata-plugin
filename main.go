package main

import (
	"code.cloudfoundry.org/cli/plugin"
	"encoding/json"
	"fmt"
	. "github.com/zrob/cfmetadata-plugin/util"
	"strings"
)

type CFMetadataPlugin struct{}

type ResourceList struct {
	Resources []ResourceModel `json:"resources"`
}

type ResourceModel struct {
	Guid     string        `json:"guid,omitempty"`
	Metadata MetadataModel `json:"metadata"`
}

type MetadataModel struct {
	Labels      map[string]*string `json:"labels"`
	Annotations map[string]*string `json:"annotations"`
}

func main() {
	plugin.Start(new(CFMetadataPlugin))
}

func (c *CFMetadataPlugin) Run(cliConnection plugin.CliConnection, args []string) {
	argCount := len(args)

	if args[0] == "annotations" {
		if argCount < 3 {
			fmt.Println(c.GetMetadata().Commands[0].UsageDetails.Usage)
		} else if argCount > 3 {
			c.setAnnotations(cliConnection, args[1:])
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
					Usage: "cf annotations RESOURCE RESOURCE_NAME KEY=VAL",
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

	response := stringifyCurlResponse(output)
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

	displayAnnotations(resources.Resources[0], resource, name)
}

func (c *CFMetadataPlugin) setAnnotations(cliConnection plugin.CliConnection, args []string) {
	resource := args[0]
	name := args[1]
	annotationsToAdd := make(map[string]string)
	var annotationsToRemove []string

	for _, a := range args[2:] {
		if strings.Contains(a, "=") {
			annotation := strings.Split(a, "=")
			if len(annotation) != 2 {
				fmt.Println("Annotations must be in the format of KEY=VAL or KEY-")
				return
			}
			annotationsToAdd[annotation[0]] = annotation[1]
		} else if strings.HasSuffix(a, "-") {
			annotationsToRemove = append(annotationsToRemove, strings.TrimSuffix(a, "-"))
		} else {
			fmt.Println("Annotations must be in the format of KEY=VAL or KEY-")
			return
		}
	}

	output, err := cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("v3/%ss?names=%s", resource, name))
	FreakOut(err)

	response := stringifyCurlResponse(output)
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

	url := fmt.Sprintf("v3/%ss/%s", resource, resources.Resources[0].Guid)

	entityToAdd := ResourceModel{}
	entityToAdd.Metadata.Annotations = make(map[string]*string)

	for key, val := range annotationsToAdd {
		localVal := val
		entityToAdd.Metadata.Annotations[key] = &localVal
	}
	for _, key := range annotationsToRemove {
		entityToAdd.Metadata.Annotations[key] = nil
	}
	updateRequest, err := json.Marshal(entityToAdd)

	output, err = cliConnection.CliCommandWithoutTerminalOutput("curl", url, "-X", "PATCH", "-d", string(updateRequest))
	FreakOut(err)

	entity := ResourceModel{}
	response = stringifyCurlResponse(output)
	err = json.Unmarshal([]byte(response), &entity)
	FreakOut(err)

	displayAnnotations(entity, resource, name)
}

func stringifyCurlResponse(output []string) string {
	var responseString string
	for _, part := range output {
		responseString += part
	}
	return responseString
}

func displayAnnotations(entity ResourceModel, resource string, name string) {
	fmt.Printf("Annotations for %s %s\r\n\r\n", resource, name)
	if len(entity.Metadata.Annotations) == 0 {
		fmt.Println("None")
	} else {
		for key, val := range entity.Metadata.Annotations {
			fmt.Printf("%s: %s\r\n", key, *val)
		}
	}
}
