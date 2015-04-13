package main

import (
	"encoding/json"
	"fmt"
	"github.com/Grayda/go-orvibo"
	"github.com/ninjasphere/go-ninja/model"
	"github.com/ninjasphere/go-ninja/suit"
)

type configService struct {
	driver *OrviboDriver
}

func (c *configService) GetActions(request *model.ConfigurationRequest) (*[]suit.ReplyAction, error) {
	return &[]suit.ReplyAction{
		suit.ReplyAction{
			Name:  "",
			Label: "Orvibo AllOnes",
		},
	}, nil
}

func (c *configService) Configure(request *model.ConfigurationRequest) (*suit.ConfigurationScreen, error) {
	fmt.Sprintf("Incoming configuration request. Action:%s Data:%s", request.Action, string(request.Data))

	switch request.Action {
	case "list":
		return c.list()
	case "":
		if len(device) > 0 {
			return c.list()
		}
		fallthrough
	case "learn":

		var vals map[string]string
		json.Unmarshal(request.Data, &vals)

		orvibo.EnterLearningMode(vals["allone"])

		return c.list()

	default:
		return c.error(fmt.Sprintf("Unknown action: %s", request.Action))
	}
}

func (c *configService) error(message string) (*suit.ConfigurationScreen, error) {

	return &suit.ConfigurationScreen{
		Sections: []suit.Section{
			suit.Section{
				Contents: []suit.Typed{
					suit.Alert{
						Title:        "Error",
						Subtitle:     message,
						DisplayClass: "danger",
					},
				},
			},
		},
		Actions: []suit.Typed{
			suit.ReplyAction{
				Label: "Cancel",
				Name:  "list",
			},
		},
	}, nil
}
func (c *configService) list() (*suit.ConfigurationScreen, error) {

	var allones []suit.ActionListOption

	for _, allone := range device {
		allones = append(allones, suit.ActionListOption{
			Title: allone.Device.Name,
			//Subtitle: tv.ID,
			Value: allone.Device.MACAddress,
		})
	}

	screen := suit.ConfigurationScreen{
		Title: "Orvibo AllOnes",
		Sections: []suit.Section{
			suit.Section{
				Contents: []suit.Typed{
					suit.ActionList{
						Name:    "allone",
						Options: allones,
						PrimaryAction: &suit.ReplyAction{
							Name:        "learn",
							DisplayIcon: "pencil",
						},
					},
				},
			},
		},
		Actions: []suit.Typed{
			suit.CloseAction{
				Label: "Close",
			},
			suit.ReplyAction{
				Label:        "New TV",
				Name:         "new",
				DisplayClass: "success",
				DisplayIcon:  "star",
			},
		},
	}

	return &screen, nil
}

/*
func (c *configService) edit(config TVConfig) (*suit.ConfigurationScreen, error) {

	title := "New Samsung TV"
	if config.ID != "" {
		title = "Editing Samsung TV"
	}

	screen := suit.ConfigurationScreen{
		Title: title,
		Sections: []suit.Section{
			suit.Section{
				Contents: []suit.Typed{
					suit.InputHidden{
						Name:  "id",
						Value: config.ID,
					},
					suit.InputText{
						Name:        "name",
						Before:      "Name",
						Placeholder: "My TV",
						Value:       config.Name,
					},
					suit.InputText{
						Name:        "host",
						Before:      "Host",
						Placeholder: "IP or Hostname",
						Value:       config.Host,
					},
				},
			},
		},
		Actions: []suit.Typed{
			suit.CloseAction{
				Label: "Cancel",
			},
			suit.ReplyAction{
				Label:        "Save",
				Name:         "save",
				DisplayClass: "success",
				DisplayIcon:  "star",
			},
		},
	}

	return &screen, nil
}
*/
func i(i int) *int {
	return &i
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
