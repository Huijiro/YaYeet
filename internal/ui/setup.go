package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/Huijiro/YaYeet/internal/config"
	"github.com/Huijiro/YaYeet/internal/runner"
)

func setupPage(configuration *config.Configuration, runners []runner.Runner, window fyne.Window, actionLabel string, onComplete func()) fyne.CanvasObject {
	runnerOptions := make([]string, 0, len(runners))
	runnerByOption := make(map[string]runner.Runner, len(runners))
	for _, discovered := range runners {
		option := discovered.Name + " (" + discovered.Version + ", " + string(discovered.Source) + ")"
		runnerOptions = append(runnerOptions, option)
		runnerByOption[option] = discovered
	}

	runnerSelect := widget.NewSelect(runnerOptions, func(option string) {
		if discovered, ok := runnerByOption[option]; ok {
			configuration.SetRunner(discovered)
		}
	})

	selectedRunner := configuration.Runner.Executable
	if selectedRunner == "" && len(runners) > 0 {
		defaultRunner := config.DefaultRunner(runners)
		configuration.SetRunner(defaultRunner)
		selectedRunner = defaultRunner.Executable
	}
	for option, discovered := range runnerByOption {
		if discovered.Executable == selectedRunner {
			runnerSelect.SetSelected(option)
			break
		}
	}

	installationPath := widget.NewEntry()
	installationPath.SetText(configuration.InstallationPath)

	installationPicker := widget.NewButton("Choose folder", func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err == nil && uri != nil {
				installationPath.SetText(uri.Path())
			}
		}, window)
	})

	prefixPath := widget.NewEntry()
	prefixPath.SetText(configuration.WinePrefix)

	prefixPicker := widget.NewButton("Choose folder", func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err == nil && uri != nil {
				prefixPath.SetText(uri.Path())
			}
		}, window)
	})

	installationRow := container.New(layout.NewFormLayout(),
		widget.NewLabel("Game installation"),
		container.NewBorder(nil, nil, nil, installationPicker, installationPath),
	)
	prefixRow := container.New(layout.NewFormLayout(),
		widget.NewLabel("Wine/Proton prefix"),
		container.NewBorder(nil, nil, nil, prefixPicker, prefixPath),
	)

	status := widget.NewLabel("")
	continueButton := widget.NewButton(actionLabel, func() {
		configuration.InstallationPath = installationPath.Text
		configuration.WinePrefix = prefixPath.Text
		if err := configuration.Save(); err != nil {
			status.SetText("Could not save configuration: " + err.Error())
			return
		}
		onComplete()
	})

	content := container.NewVBox(
		widget.NewLabel("Choose where the game and its Wine prefix should live. Leave the options as is to use defaults."),
		installationRow,
		prefixRow,
		container.NewBorder(nil, nil, widget.NewLabel("Wine/Proton runner"), nil, runnerSelect),
		continueButton,
		status,
	)

	return container.NewCenter(content)
}
