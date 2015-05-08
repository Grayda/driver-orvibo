package main

import (
	"encoding/json"
	"fmt"
	"github.com/Grayda/go-orvibo"
	"github.com/ninjasphere/go-ninja/model"
	"github.com/ninjasphere/go-ninja/suit"
	"strings"
)

// This file contains most of the code for the UI (i.e. what appears in the Labs)

type configService struct {
	driver *OrviboDriver
}

// This function is common across all UIs, and is called by the Sphere. Shows our menu option on the main Labs screen
func (c *configService) GetActions(request *model.ConfigurationRequest) (*[]suit.ReplyAction, error) {
	// What we're going to show
	var screen []suit.ReplyAction
	// Loop through all Orvibo devices. All menu options lead to the same page anyway
	for _, allone := range driver.device {
		// If it's an AllOne
		if allone.Device.DeviceType == orvibo.ALLONE {
			// Add a menu option
			screen = append(screen, suit.ReplyAction{
				Name:        "",
				Label:       "Configure AllOne",
				DisplayIcon: "play",
			},
			)
			break // Who cares how many AllOnes we've found? One is enough to show the UI
		}
	}
	// Return our screen for rendering
	return &screen, nil
}

// When you click on a ReplyAction button, Configure is called
func (c *configService) Configure(request *model.ConfigurationRequest) (*suit.ConfigurationScreen, error) {
	fmt.Sprintf("Incoming configuration request. Action:%s Data:%s", request.Action, string(request.Data))

	switch request.Action {
	case "list":
		fmt.Println("Showing list of IR codes..")
		return c.list()
	case "blastir":

		var vals map[string]string
		json.Unmarshal(request.Data, &vals)

		var codes = strings.Split(vals["code"], "|")
		fmt.Println("Blasting IR code " + codes[0] + " on AllOne: " + codes[1] + "..")
		orvibo.EmitIR(codes[0], codes[1])
		return c.list()
	case "new":
		return c.new(driver.config)
	case "reset": // For debugging purposes. Clears out the stored codes
		driver.config.Codes = nil
		driver.config.learningIR = false
		driver.config.learningIRName = ""
		driver.SendEvent("config", driver.config)
		return c.list()
	case "delete":
		var vals map[string]string
		err := json.Unmarshal(request.Data, &vals)
		if err != nil {
			return c.error(fmt.Sprintf("Failed to unmarshal save config request %s: %s", request.Data, err))
		}
		var codes = strings.Split(vals["code"], "|")
		driver.deleteIR(driver.config, codes[0])

		return c.list()

	case "save":
		var vals map[string]string
		err := json.Unmarshal(request.Data, &vals)
		if err != nil {
			return c.error(fmt.Sprintf("Failed to unmarshal save config request %s: %s", request.Data, err))
		}

		driver.config.learningIR = true
		driver.config.learningIRName = vals["name"]
		driver.config.learningIRDescription = vals["description"]
		driver.config.learningIRDevice = vals["allone"]
		orvibo.EnterLearningMode(vals["allone"])

		return c.confirm("Learning IR code", "Please press a button on your remote. Click 'Okay' when done")
	case "":
		return c.list()

		fallthrough

	default:

		// return c.list()
		return c.error(fmt.Sprintf("Unknown action: %s", request.Action))
	}
	return nil, nil
}

func (c *configService) confirm(title string, description string) (*suit.ConfigurationScreen, error) {
	screen := suit.ConfigurationScreen{
		Title: title,
		Sections: []suit.Section{
			suit.Section{
				Contents: []suit.Typed{
					suit.StaticText{
						Title: "About this screen",
						Value: description,
					},
				},
			},
		},
		Actions: []suit.Typed{
			suit.ReplyAction{
				Label:        "Okay",
				Name:         "list",
				DisplayClass: "success",
				DisplayIcon:  "ok",
			},
		},
	}

	return &screen, nil
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
				Label:        "Cancel",
				Name:         "list",
				DisplayClass: "success",
				DisplayIcon:  "ok",
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
			Value:    code.Code + "|" + code.AllOne,
		})
	}

	screen := suit.ConfigurationScreen{
		Title: "Saved IR Codes",
		Sections: []suit.Section{
			suit.Section{
				Contents: []suit.Typed{
					suit.StaticText{
						Title: "About this screen",
						Value: "This page shows saved IR codes. To add a new code, press 'New IR Code'",
					},
					suit.ActionList{
						Name:    "code",
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
				Label:        "New IR Code",
				Name:         "new",
				DisplayClass: "success",
				DisplayIcon:  "asterisk",
			},
		},
	}

	return &screen, nil
}

func (c *configService) new(config *OrviboDriverConfig) (*suit.ConfigurationScreen, error) {

	// What we're going to show
	var allones []suit.RadioGroupOption

	allones = append(allones, suit.RadioGroupOption{
		Title:       "All Connected AllOnes",
		Value:       "ALL",
		DisplayIcon: "globe",
	})

	// Loop through all Orvibo devices. All menu options lead to the same page anyway
	for _, allone := range driver.device {
		// If it's an AllOne
		if allone.Device.DeviceType == orvibo.ALLONE {
			// Add a menu option
			allones = append(allones, suit.RadioGroupOption{
				Title:       allone.Device.Name,
				DisplayIcon: "play",
				Value:       allone.Device.MACAddress,
			},
			)

		}
	}

	title := "New IR Code"

	screen := suit.ConfigurationScreen{
		Title: title,
		Sections: []suit.Section{
			suit.Section{
				Contents: []suit.Typed{
					suit.StaticText{
						Title: "About this screen",
						Value: "Please enter a name and a description for this code. You must also pick an AllOne. When you're ready, click 'Start Learning' and press a button on your remote",
					},
					suit.InputHidden{
						Name:  "id",
						Value: "",
					},
					suit.InputText{
						Name:        "name",
						Before:      "Name for this code",
						Placeholder: "TV On",
						Value:       "",
					},
					suit.InputText{
						Name:        "description",
						Before:      "Code Description",
						Placeholder: "Living Room TV On",
						Value:       "",
					},
					suit.RadioGroup{
						Title:   "Select an AllOne to blast from",
						Name:    "allone",
						Options: allones,
					},
				},
			},
		},
		Actions: []suit.Typed{
			suit.ReplyAction{
				Label:        "Cancel",
				Name:         "list",
				DisplayClass: "default",
			},
			suit.ReplyAction{
				Label:        "Start Learning",
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
