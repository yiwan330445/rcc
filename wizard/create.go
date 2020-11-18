package wizard

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/manifoldco/promptui"
	"github.com/robocorp/rcc/common"
	"github.com/robocorp/rcc/operations"
	"github.com/robocorp/rcc/pathlib"
	"github.com/robocorp/rcc/pretty"
)

func Create(arguments []string) error {
	common.Stdout("\n")

	warning(len(arguments) > 1, "You provided more than one argument, but only the first one will be\nused as the name.")
	prompt := promptui.Prompt{
		Label:    "Give robot name",
		Default:  firstOf(arguments, "my-first-robot"),
		Validate: hasLength,
	}
	robotName, err := prompt.Run()
	common.Stdout("\n")

	if err != nil {
		return err
	}

	fullpath, err := filepath.Abs(robotName)
	if err != nil {
		return err
	}

	if pathlib.IsDir(fullpath) {
		return errors.New(fmt.Sprintf("Folder %s already exists. Try with other name.", robotName))
	}

	selection := promptui.Select{
		Label: "Choose a template",
		Items: operations.ListTemplates(),
	}

	_, selected, err := selection.Run()
	common.Stdout("\n")

	if err != nil {
		return err
	}

	common.Stdout("%sCreating the %s%s%s robot: %s%s%s\n", pretty.White, pretty.Cyan, selected, pretty.White, pretty.Cyan, robotName, pretty.Reset)
	common.Stdout("\n")

	err = operations.InitializeWorkarea(fullpath, selected, false)
	if err != nil {
		return err
	}

	common.Stdout("%sThe %s robot has been created to: %s%s\n", pretty.Green, selected, robotName, pretty.Reset)
	common.Stdout("\n")

	common.Stdout("%sGet started with following commands:%s\n", pretty.White, pretty.Reset)
	common.Stdout("\n")

	common.Stdout("%s$ %scd %s%s\n", pretty.Grey, pretty.Cyan, robotName, pretty.Reset)
	common.Stdout("%s$ %srcc run%s\n", pretty.Grey, pretty.Cyan, pretty.Reset)
	common.Stdout("\n")

	return nil
}
