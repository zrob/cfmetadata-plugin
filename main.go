package main

import (
	"code.cloudfoundry.org/cli/plugin"
	"encoding/json"
	"errors"
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
	if args[0] == "labels" {
		if argCount < 3 {
			fmt.Println(c.GetMetadata().Commands[1].UsageDetails.Usage)
		} else if argCount > 3 {
			c.setLabels(cliConnection, args[1:])
		} else {
			c.getLabels(cliConnection, args[1:])
		}
	}
	if args[0] == "metadata" {
		if argCount != 3 {
			fmt.Println(c.GetMetadata().Commands[2].UsageDetails.Usage)
		} else {
			c.getMetadata(cliConnection, args[1:])
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
					Usage: "cf annotations RESOURCE RESOURCE_NAME KEY=VAL KEY-",
				},
			},
			{
				Name:     "labels",
				HelpText: "view or modify labels for an API resource",
				UsageDetails: plugin.Usage{
					Usage: "cf labels RESOURCE RESOURCE_NAME KEY=VAL KEY-",
				},
			},
			{
				Name:     "metadata",
				HelpText: "view labels and annotations for an API resource",
				UsageDetails: plugin.Usage{
					Usage: "cf metadata RESOURCE RESOURCE_NAME",
				},
			},
		},
	}
}

func (c *CFMetadataPlugin) getMetadata(cliConnection plugin.CliConnection, args []string) {
	resource := args[0]
	name := args[1]

	entity, err := fetchResourceByName(cliConnection, resource, name)
	FreakOut(err)

	displayLabels(entity, resource, name)
	fmt.Println("")
	displayAnnotations(entity, resource, name)
}

func (c *CFMetadataPlugin) getAnnotations(cliConnection plugin.CliConnection, args []string) {
	resource := args[0]
	name := args[1]

	entity, err := fetchResourceByName(cliConnection, resource, name)
	FreakOut(err)

	displayAnnotations(entity, resource, name)
}

func (c *CFMetadataPlugin) setAnnotations(cliConnection plugin.CliConnection, args []string) {
	resource := args[0]
	name := args[1]

	annotationsToAdd, annotationsToRemove, err := parseSetUnsetArguments(args[2:], "Annotations must be in the format of KEY=VAL or KEY-")
	FreakOut(err)

	currentEntity, err := fetchResourceByName(cliConnection, resource, name)
	FreakOut(err)

	updateEntity := ResourceModel{}
	updateEntity.Metadata.Annotations = make(map[string]*string)
	for key, val := range annotationsToAdd {
		localVal := val
		updateEntity.Metadata.Annotations[key] = &localVal
	}
	for _, key := range annotationsToRemove {
		updateEntity.Metadata.Annotations[key] = nil
	}

	resultEntity, err := updateResource(cliConnection, updateEntity, resource, currentEntity.Guid)
	FreakOut(err)

	displayAnnotations(resultEntity, resource, name)
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

func (c *CFMetadataPlugin) getLabels(cliConnection plugin.CliConnection, args []string) {
	resource := args[0]
	name := args[1]

	entity, err := fetchResourceByName(cliConnection, resource, name)
	FreakOut(err)

	displayLabels(entity, resource, name)
}

func (c *CFMetadataPlugin) setLabels(cliConnection plugin.CliConnection, args []string) {
	resource := args[0]
	name := args[1]

	labelsToAdd, labelsToRemove, err := parseSetUnsetArguments(args[2:], "Labels must be in the format of KEY=VAL or KEY-")
	FreakOut(err)

	currentEntity, err := fetchResourceByName(cliConnection, resource, name)
	FreakOut(err)

	updateEntity := ResourceModel{}
	updateEntity.Metadata.Labels = make(map[string]*string)
	for key, val := range labelsToAdd {
		localVal := val
		updateEntity.Metadata.Labels[key] = &localVal
	}
	for _, key := range labelsToRemove {
		updateEntity.Metadata.Labels[key] = nil
	}

	resultEntity, err := updateResource(cliConnection, updateEntity, resource, currentEntity.Guid)
	FreakOut(err)

	displayLabels(resultEntity, resource, name)
}

func displayLabels(entity ResourceModel, resource string, name string) {
	fmt.Printf("Labels for %s %s\r\n\r\n", resource, name)
	if len(entity.Metadata.Labels) == 0 {
		fmt.Println("None")
	} else {
		for key, val := range entity.Metadata.Labels {
			fmt.Printf("%s: %s\r\n", key, *val)
		}
	}
}

func stringifyCurlResponse(output []string) string {
	var responseString string
	for _, part := range output {
		responseString += part
	}
	return responseString
}

func parseSetUnsetArguments(args []string, errorText string) (toAdd map[string]string, toRemove []string, err error) {
	toAdd = make(map[string]string)

	for _, arg := range args {
		if strings.Contains(arg, "=") {
			addArgPieces := strings.Split(arg, "=")

			if len(addArgPieces) != 2 {
				err = errors.New(errorText)
				return
			}

			toAdd[addArgPieces[0]] = addArgPieces[1]
		} else if strings.HasSuffix(arg, "-") {
			toRemove = append(toRemove, strings.TrimSuffix(arg, "-"))
		} else {
			err = errors.New(errorText)
			return
		}
	}

	return
}

func fetchResourceByName(cliConnection plugin.CliConnection, resource string, name string) (entity ResourceModel, err error) {
	output, err := cliConnection.CliCommandWithoutTerminalOutput("curl", fmt.Sprintf("v3/%ss?names=%s", resource, name))
	if err != nil {
		return
	}

	response := stringifyCurlResponse(output)
	resources := ResourceList{}
	err = json.Unmarshal([]byte(response), &resources)
	if err != nil {
		return
	}

	if len(resources.Resources) == 0 {
		err = errors.New(fmt.Sprintf("%s %s not found\r\n", resource, name))
		return
	} else if len(resources.Resources) > 1 {
		err = errors.New(fmt.Sprintf("%s %s is ambiguous, more than one result returned\r\n", resource, name))
		return
	}

	entity = resources.Resources[0]
	return
}

func updateResource(cliConnection plugin.CliConnection, updateEntity ResourceModel, resource string, guid string) (resultEntity ResourceModel, err error) {
	updateUrl := fmt.Sprintf("v3/%ss/%s", resource, guid)
	updateRequest, err := json.Marshal(updateEntity)
	if err != nil {
		return
	}

	output, err := cliConnection.CliCommandWithoutTerminalOutput("curl", updateUrl, "-X", "PATCH", "-d", string(updateRequest))
	if err != nil {
		return
	}

	response := stringifyCurlResponse(output)
	err = json.Unmarshal([]byte(response), &resultEntity)

	return
}
