package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Grayda/go-orvibo"
	"github.com/ninjasphere/go-ninja/model"
	"github.com/ninjasphere/go-ninja/suit"
)

// This file contains most of the code for the UI (i.e. what appears in the Labs)

type configService struct {
	driver *OrviboDriver
}

// This function is common across all UIs, and is called by the Sphere. Shows our menu option on the main Labs screen
// The "c" bit at the start means this func is an extension of the configService struct (like prototyping, I think?)
func (c *configService) GetActions(request *model.ConfigurationRequest) (*[]suit.ReplyAction, error) {
	// What we're going to show
	var screen []suit.ReplyAction
	// Loop through all Orvibo devices. We do this so we can find an AllOne
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
			// We've found at least one AllOne. That's enough to show the UI, so we stop looping
			break
		}
	}
	// Return our screen to the sphere-ui for rendering
	return &screen, nil
}

// When you click on a ReplyAction button (e.g. the "Configure AllOne" button defined above), Configure is called. requests.Action == the "Name" portion of the ReplyAction
func (c *configService) Configure(request *model.ConfigurationRequest) (*suit.ConfigurationScreen, error) {
	fmt.Sprintf("Incoming configuration request. Action:%s Data:%s", request.Action, string(request.Data))

	switch request.Action {
	case "list": // Listing the IR codes
		fmt.Println("Showing list of IR codes..")
		return c.list()
	case "blastir": // Blasting IR codes
		// Make a map of strings
		var vals map[string]string
		// Take our json response from sphere-ui and place it into our vals map
		json.Unmarshal(request.Data, &vals)
		// Because ActionListOption can't return more than one lot of data and we need to let our code know
		// WHAT code to blast, and what AllOne to shoot it from, we use a pipe to mash data together
		var codes = strings.Split(vals["code"], "|")
		fmt.Println("Blasting IR code " + codes[0] + " on AllOne: " + codes[1] + "..")
		// As inferred from the fmt.Println above, codes[0] (being vals["code"] split by the | command) is the IR and codes[1] is the AllOne to shoot from (MAC Address)
		orvibo.EmitIR(codes[0], codes[1])
		// c.list creates a list of AllOne IR codes and sends them back to sphere-ui / suits for displaying
		return c.list()
	case "new": // If we've clicked the New IR button
		// Returns a configuration screen with textboxes and stuff, to allow users to set up a new IR code
		return c.new(driver.config)
	case "reset": // For debugging purposes. Clears out the stored codes
		driver.config.Codes = nil
		driver.config.learningIR = false
		driver.config.learningIRName = ""
		driver.SendEvent("config", driver.config) // Writes the changes back to config
		return c.list()
	case "delete": // Delete a code. Very similar to the blastIR code above. Takes the reply, splits it by "|" and then passes that to driver.deleteIR
		var vals map[string]string
		err := json.Unmarshal(request.Data, &vals)
		if err != nil {
			return c.error(fmt.Sprintf("Failed to unmarshal save config request %s: %s", request.Data, err))
		}
		var codes = strings.Split(vals["code"], "|")
		driver.deleteIR(driver.config, codes[0])

		return c.list() // Take us back to the list of saved IR codes
	case "newgroup": // Similar to "new", but takes us to a group creation page
		return c.newgroup(driver.config)
	case "savegroup": // We've hit the "Save" button on the "new group" page. Time to save the options!
		var vals map[string]string
		err := json.Unmarshal(request.Data, &vals)
		if err != nil {
			return c.error(fmt.Sprintf("Failed to unmarshal save config request %s: %s", request.Data, err))
		}

		// Add our new group to the end of the "CodeGroups" in the driver's configuration.
		driver.config.CodeGroups = append(driver.config.CodeGroups, OrviboIRCodeGroup{
			Name:        vals["name"],
			Description: vals["description"],
		})
		driver.saveGroups(driver.config)
		return c.list()
	case "save": // Very similar to savegroup, but saves an IR code instead
		var vals map[string]string
		err := json.Unmarshal(request.Data, &vals)
		if err != nil {
			return c.error(fmt.Sprintf("Failed to unmarshal save config request %s: %s", request.Data, err))
		}

		// Now we tell our driver we're being put into learning mode.
		driver.config.learningIR = true
		driver.config.learningIRName = vals["name"]
		driver.config.learningIRDescription = vals["description"]
		driver.config.learningIRDevice = vals["allone"]
		driver.config.learningIRGroup = vals["group"]
		// Tell the driver to put vals["allone"] (the MAC Address of our AllOne) into learning mode. Give it the MAC Address "ALL" to put All AllOnes into learning mode
		orvibo.EnterLearningMode(vals["allone"])

		// The UI isn't event driven, meaning we can't tell the UI to pause until we get an IR code back. If we go back to c.list(), there won't be a code there (because we're
		// still learning), so we shove this page in the middle that makes the user click OK when done. When they do, the code has already been learned and shows up in the UI
		return c.confirm("Learning IR code", "Please press a button on your remote. Click 'Okay' when done")
	case "": // Coming in from the main menu
		return c.list()

	default: // Everything else

		// return c.list()
		return c.error(fmt.Sprintf("Unknown action: %s", request.Action))
	}

	// If this code runs, then we done fucked up, because default: didn't catch. When this code runs, the universe melts into a gigantic heap. But
	// removing this violates Apple guidelines and ensures the downfall of humanity (probably) so I don't want to risk it.
	// Then again, I could be making all this up. Do you want to remove it and try? ( ͡° ͜ʖ ͡°)
	return nil, nil
}

// So this function (which is an extension of the configService struct that suit (or Sphere-UI) requires) creates a box with a single "Okay" button and puts in a title and text
func (c *configService) confirm(title string, description string) (*suit.ConfigurationScreen, error) {
	// We create a new suit.ConfigurationScreen which is a whole page within the UI
	screen := suit.ConfigurationScreen{
		Title: title,
		Sections: []suit.Section{ // The UI lets us create sections for separating options. This line creates an array of sections
			suit.Section{ // And within that array of sections, a single section
				Contents: []suit.Typed{ // The contents of that section. I don't know what suit.Typed is. It's an interface, but asides from that, I don't know much else just yet
					suit.StaticText{ // Create some static text
						Title: "About this screen",
						Value: description,
					},
				},
			},
		},
		Actions: []suit.Typed{ // This configuration screen can show actionable buttons at the bottom. ReplyAction, as shown above, calls Configure. There is also CloseAction for cancel buttons
			suit.ReplyAction{
				Label:        "Okay",
				Name:         "list",
				DisplayClass: "success", // These are bootstrap classes (or rather, font-awesome classes). They are basically btn-*, where * is DisplayClass (e.g. btn-success)
				DisplayIcon:  "ok",      // Same as above. If you want to show fa-open-folder, you'd set DisplayIcon to "open-folder"
			},
		},
	}

	return &screen, nil
}

// Error! Same as above. It's a function that is added on to configService and displays an error message
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
			suit.ReplyAction{ // Shows a button we can click on. Takes us back to c.Configuration (reply.Action will be "list")
				Label:        "Cancel",
				Name:         "list",
				DisplayClass: "success",
				DisplayIcon:  "ok",
			},
		},
	}, nil
}

// The meat of our UI. Shows a list of IR codes to be blasted. This could show anything you like, really.
func (c *configService) list() (*suit.ConfigurationScreen, error) {

	// ActionListOption are buttons within a section that are sent to c.Configuration
	var codes []suit.ActionListOption
	// Sections, for logical grouping
	var sections []suit.Section
	// Loop through all the CodeGroups in our driver
	for _, groups := range driver.config.CodeGroups {
		// Go through all the saved IR codes
		for _, code := range driver.config.Codes {
			// If this IR code belongs to the group we're iterating through
			if code.Group == groups.Name {
				// Add this code to the section
				codes = append(codes, suit.ActionListOption{
					Title:    code.Name,
					Subtitle: code.Description,
					Value:    code.Code + "|" + code.AllOne, // We need to add a pipe because we can't set Value: twice, so we mash two lots of data together, split up by a "|"
				})

			} // End If

		} // Loop through the next code

		// Now that we've looped through the codes for this group, create our UI for that section
		sections = append(sections, suit.Section{ // Append a new suit.Section into our sections variable

			Contents: []suit.Typed{ // Again, dunno what this means
				suit.StaticText{ // Create static text (a heading, basically)
					Title: groups.Name,        // With the name of the IR code group
					Value: groups.Description, // And it's description
				},
				suit.ActionList{ // Now, the buttons we can click to emit IR!
					Name:    "code", // This doesn't get sent back to c.Configuration, only PrimaryAction and SecondaryAction do. Tricky, eh?
					Options: codes,  // The options we can click on, is the map of ActionListOption we created above (where we had to stick in our pipe) (hey, phrasing!)
					PrimaryAction: &suit.ReplyAction{ // This is the main button. It could be a cancel button for all we care. It's a primary action button.
						Name:         "blastir", // Similar to above. As it's a ReplyAction, it gets sent to c.Configuration. Notice a pattern here?
						Label:        "Blast",   // The text that appears on the button
						DisplayIcon:  "star",
						DisplayClass: "danger",
					},
					SecondaryAction: &suit.ReplyAction{ // Secondary buttons appear alongside the primary button, but they're smaller (e.g. [          PRIMARY BUTTON          ][ SECONDARY ])
						Name:         "delete",
						Label:        "Delete",
						DisplayIcon:  "trash",
						DisplayClass: "danger",
					},
				},
			},
		},
		)
		codes = nil // We need to empty out our codes array, otherwise the next section will contain codes from the first group, in addition to the second group
	}

	// Now that we've looped and got our sections, it's time to build the actual screen
	screen := suit.ConfigurationScreen{
		Title:    "Saved IR Codes",
		Sections: sections, // Our sections. Contains all the buttons and everything!
		Actions: []suit.Typed{ // Actiosn you can take on this page
			suit.CloseAction{ // Here we go! This takes a label and probably a DisplayIcon and DisplayClass and just takes you back to the main screen. Not YOUR main screen though, so use a ReplyAction with a "" name to go back to YOUR menu
				Label: "Close",
			},
			suit.ReplyAction{ // Reply action. Same as the rest
				Label:        "New IR Code",
				Name:         "new", // Back in c.Configuration, show the new code UI
				DisplayClass: "success",
				DisplayIcon:  "asterisk",
			},
			suit.ReplyAction{ // You can have as many ReplyActions as you like (I think) and it'll squeeze them in side by side
				Label:        "New IR Group",
				Name:         "newgroup",
				DisplayClass: "default",
				DisplayIcon:  "asterisk",
			},
		},
	}

	return &screen, nil
}

// Shows the UI to learn a new IR code
func (c *configService) new(config *OrviboDriverConfig) (*suit.ConfigurationScreen, error) {

	// What radio options we're going to show. Are you seeing a pattern here now? The UI is rather easy once you do it for a while
	// If you want to know what options the UI supports, and what values you can use with them, check out https://github.com/ninjasphere/go-ninja/blob/master/suit/screen.go
	var allones []suit.RadioGroupOption
	var groups []suit.RadioGroupOption

	// Add a new RadioGroupOption to our list. This one blasts from All AllOnes connected ("ALL" is a special MAC Address in go-orvibo)
	allones = append(allones, suit.RadioGroupOption{
		Title:       "All Connected AllOnes",
		Value:       "ALL",
		DisplayIcon: "globe",
	})

	// Loop through the groups we've got
	for _, codegroup := range driver.config.CodeGroups {
		groups = append(groups, suit.RadioGroupOption{ // Add a new radio buton
			Title:       codegroup.Name,
			Value:       codegroup.Name,
			DisplayIcon: "folder-open",
		},
		)
	}

	// Loop through all Orvibo devices.
	for _, allone := range driver.device {
		// If it's an AllOne
		if allone.Device.DeviceType == orvibo.ALLONE {
			// Add a Radio button with our AllOne's name and MAC Address
			allones = append(allones, suit.RadioGroupOption{
				Title:       allone.Device.Name,
				DisplayIcon: "play",
				Value:       allone.Device.MACAddress,
			},
			)

		}
	}

	title := "New IR Code" // Up here for readability

	screen := suit.ConfigurationScreen{
		Title: title,
		Sections: []suit.Section{ // New array of sections
			suit.Section{ // New section
				Contents: []suit.Typed{
					suit.StaticText{ // Some introductory text
						Title: "About this screen",
						Value: "Please enter a name and a description for this code. You must also pick an AllOne. When you're ready, click 'Start Learning' and press a button on your remote",
					},
					suit.InputHidden{ // Not actually used by my code, but you can use InputHidden to pass stuff back to c.Configure()
						Name:  "id",
						Value: "",
					},
					suit.InputText{ // Textbox
						Name:        "name",
						Before:      "Name for this code",
						Placeholder: "TV On", // Placeholder is the faded text that appears inside a textbox, giving you a hint as to what to type in
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
						Options: allones, // We created our radio group before, and now we put it in here
					},
					suit.RadioGroup{
						Title:   "Select a group to add this code to",
						Name:    "group",
						Options: groups, // Same with our code groups
					},
				},
			},
		},
		Actions: []suit.Typed{
			suit.ReplyAction{ // This is not a CloseAction, because we want to go back to the list of IR codes, not back to the main menu. Hence why we use a ReplyAction with "list"
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

// You know the drill. I don't think it even needs to accept an *OrviboDriverConfig, because you could just call driver.config
func (c *configService) newgroup(config *OrviboDriverConfig) (*suit.ConfigurationScreen, error) {

	title := "New Code Group"
	// New screen
	screen := suit.ConfigurationScreen{
		Title: title,
		Sections: []suit.Section{
			suit.Section{
				Contents: []suit.Typed{
					suit.StaticText{
						Title: "About this screen",
						Value: "On this page you can create a new group to put your codes in. For example, you might create a group called 'Living Room' to store codes relating to your home theater in your living room",
					},
					suit.InputHidden{
						Name:  "id",
						Value: "",
					},
					suit.InputText{
						Name:        "name",
						Before:      "Name for this group",
						Placeholder: "Home Theater",
						Value:       "",
					},
					suit.InputText{
						Name:        "description",
						Before:      "Description of this group",
						Placeholder: "Codes related to the home theater",
						Value:       "",
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
				Label:        "Save Group",
				Name:         "savegroup",
				DisplayClass: "success",
				DisplayIcon:  "star",
			},
		},
	}

	return &screen, nil
}

// Aye-aye, captain.
// Not actually needed (?)
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
