package main

import (
	"encoding/json"
	"fmt"
	"github.com/Grayda/go-orvibo"
	"github.com/ninjasphere/go-ninja/model"
	"github.com/ninjasphere/go-ninja/suit"
	"strconv"
)

type configService struct {
	driver *OrviboDriver
}

func (c *configService) GetActions(request *model.ConfigurationRequest) (*[]suit.ReplyAction, error) {
	var screen []suit.ReplyAction

	for _, allone := range driver.device {
		if allone.Device.DeviceType == orvibo.ALLONE {
			screen = append(screen, suit.ReplyAction{
				Name:  "",
				Label: "Configure AllOne: " + strconv.Itoa(allone.Device.ID),
			},
			)
		}
	}
	return &screen, nil
}

func (c *configService) Configure(request *model.ConfigurationRequest) (*suit.ConfigurationScreen, error) {
	fmt.Sprintf("Incoming configuration request. Action:%s Data:%s", request.Action, string(request.Data))

	switch request.Action {
	case "list":
		return c.list()
	case "blastir":
		var vals map[string]string
		json.Unmarshal(request.Data, &vals)

		orvibo.EmitIR(vals["allone"], vals["code"])
	case "new":
		return c.new(driver.config)
	case "save":
		var vals map[string]string
		err := json.Unmarshal(request.Data, &vals)
		if err != nil {
			return c.error(fmt.Sprintf("Failed to unmarshal save config request %s: %s", request.Data, err))
		}

		driver.config.learningIR = true
		driver.config.learningIRName = vals["name"]
		orvibo.EnterLearningMode("ALL")

		return c.list()
	case "":
		return c.list()

		fallthrough

	default:

		// return c.list()
		return c.error(fmt.Sprintf("Unknown action: %s", request.Action))
	}
	return nil, nil
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

	var codes []suit.ActionListOption

	for _, code := range driver.config.Codes {
		codes = append(codes, suit.ActionListOption{
			Title:    code.Name,
			Subtitle: code.Description,
			Value:    code.Code,
		})
	}

	screen := suit.ConfigurationScreen{
		Title: "Saved IR Codes",
		Sections: []suit.Section{
			suit.Section{
				Contents: []suit.Typed{
					suit.ActionList{
						Name:    "allone",
						Options: codes,
						PrimaryAction: &suit.ReplyAction{
							Name:         "blastir",
							Label:        "Blast",
							DisplayIcon:  "star",
							DisplayClass: "danger",
						},
						SecondaryAction: &suit.ReplyAction{
							Name:         "delete",
							Label:        "Delete",
							DisplayIcon:  "trash",
							DisplayClass: "danger",
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
				Label:        "New IR code",
				Name:         "new",
				DisplayClass: "success",
				DisplayIcon:  "star",
			},
		},
	}

	return &screen, nil
}

func (c *configService) new(config *OrviboDriverConfig) (*suit.ConfigurationScreen, error) {

	title := "New IR Code"

	screen := suit.ConfigurationScreen{
		Title: title,
		Sections: []suit.Section{
			suit.Section{
				Contents: []suit.Typed{
					suit.StaticText{
						Title: "About this screen",
						Value: "When you click Save, the AllOne will enter learning mode. Please press a button on your IR remote to learn it",
					},
					suit.InputHidden{
						Name:  "id",
						Value: "",
					},
					suit.InputText{
						Name:        "name",
						Before:      "Name for this code",
						Placeholder: "TV Power On",
						Value:       "",
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
