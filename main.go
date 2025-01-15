package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"image/color"
	"io"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/google/uuid"
	"github.com/jung-kurt/gofpdf"
)

// CanvasData represents the data structure for saving/loading
type CanvasData struct {
	KeyPartners      string `json:"keyPartners"`
	KeyActivities    string `json:"keyActivities"`
	KeyResources     string `json:"keyResources"`
	ValueProposition string `json:"valueProposition"`
	CustomerRel      string `json:"customerRelationships"`
	Channels         string `json:"channels"`
	CustomerSegments string `json:"customerSegments"`
	CostStructure    string `json:"costStructure"`
	RevenueStreams   string `json:"revenueStreams"`
}

// Version represents a snapshot of the canvas
type Version struct {
	ID        string
	Timestamp time.Time
	Data      CanvasData
	Comments  []Comment
}

// Comment represents user feedback on canvas sections
type Comment struct {
	ID        string
	Section   string
	Text      string
	Author    string
	Timestamp time.Time
}

// Canvas represents the main application structure
type Canvas struct {
	// Core fields
	keyPartners      *widget.Entry
	keyActivities    *widget.Entry
	keyResources     *widget.Entry
	valueProposition *widget.Entry
	customerRel      *widget.Entry
	channels         *widget.Entry
	customerSegments *widget.Entry
	costStructure    *widget.Entry
	revenueStreams   *widget.Entry
	currentTheme     string
	autoSave         bool
	lastSaved        time.Time
	undoStack        []CanvasData
	redoStack        []CanvasData
	validator        *BusinessValidator
	progressBar      *widget.ProgressBar
	writer           fyne.Window
	window           fyne.Window // Added missing field
	versions         []Version
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("Business Canvas")

	// Create canvas with enhanced features
	canvas := &Canvas{
		keyPartners:      widget.NewMultiLineEntry(),
		keyActivities:    widget.NewMultiLineEntry(),
		keyResources:     widget.NewMultiLineEntry(),
		valueProposition: widget.NewMultiLineEntry(),
		customerRel:      widget.NewMultiLineEntry(),
		channels:         widget.NewMultiLineEntry(),
		customerSegments: widget.NewMultiLineEntry(),
		costStructure:    widget.NewMultiLineEntry(),
		revenueStreams:   widget.NewMultiLineEntry(),
		currentTheme:     "professional",
		autoSave:         true,
		validator:        NewBusinessValidator(),
		progressBar:      widget.NewProgressBar(),
	}

	canvas.window = myWindow
	// Initialize the canvas
	canvas.initialize()

	// Create toolbar
	toolbar := canvas.createToolbar()

	// Create main content
	mainContent := canvas.createMainContent()

	// Create status bar
	statusBar := canvas.createStatusBar()

	// Combine all elements
	myWindow.SetContent(container.NewBorder(toolbar, statusBar, nil, nil, mainContent))
	myWindow.Resize(fyne.NewSize(1400, 900))
	myWindow.Show()

	// Start auto-save routine
	myApp.Lifecycle().SetOnStarted(func() {
		if canvas.autoSave {
			go canvas.autoSaveRoutine()
		}
	})

	myApp.Run()
}

func (c *Canvas) initialize() {
	// Set placeholders with enhanced descriptions
	c.keyPartners.SetPlaceHolder("Who are your key partners and suppliers? What resources are you acquiring from them?")
	c.keyActivities.SetPlaceHolder("What key activities does your value proposition require?")
	c.keyResources.SetPlaceHolder("What key resources does your value proposition require?")
	c.valueProposition.SetPlaceHolder("What value do you deliver to customers? Which problems are you solving?")
	c.customerRel.SetPlaceHolder("What type of relationship does each customer segment expect?")
	c.channels.SetPlaceHolder("Through which channels do your customers want to be reached?")
	c.customerSegments.SetPlaceHolder("For whom are you creating value? Who are your most important customers?")
	c.costStructure.SetPlaceHolder("What are the most important costs inherent in your business model?")
	c.revenueStreams.SetPlaceHolder("For what value are your customers willing to pay? How would they prefer to pay?")

	// Initialize validation
	c.validator = NewBusinessValidator()

	// Set up keyboard shortcuts
	c.setupKeyboardShortcuts()

	// Set up dynamic validation
	c.setupDynamicValidation(c.keyPartners, "Key Partners")
	c.setupDynamicValidation(c.keyActivities, "Key Activities")
	c.setupDynamicValidation(c.keyResources, "Key Resources")
	c.setupDynamicValidation(c.valueProposition, "Value Proposition")
	c.setupDynamicValidation(c.customerRel, "Customer Relationships")
	c.setupDynamicValidation(c.channels, "Channels")
	c.setupDynamicValidation(c.customerSegments, "Customer Segments")
	c.setupDynamicValidation(c.costStructure, "Cost Structure")
	c.setupDynamicValidation(c.revenueStreams, "Revenue Streams")
}

func (c *Canvas) createToolbar() *widget.Toolbar {
	themeToggle := widget.NewToolbarAction(theme.ColorPaletteIcon(), func() {
		if c.currentTheme == "professional" {
			c.currentTheme = "light"
			myApp := fyne.CurrentApp()
			myApp.Settings().SetTheme(theme.LightTheme())
		} else {
			c.currentTheme = "professional"
			myApp := fyne.CurrentApp()
			myApp.Settings().SetTheme(theme.DarkTheme())
		}
	})

	saveAction := widget.NewToolbarAction(theme.DocumentSaveIcon(), func() {
		c.saveCanvas()
	})

	loadAction := widget.NewToolbarAction(theme.FolderOpenIcon(), func() {
		c.loadCanvas()
	})

	exportAction := widget.NewToolbarAction(theme.DocumentCreateIcon(), func() {
		c.exportToPDF()
	})

	validateAction := widget.NewToolbarAction(theme.ViewRefreshIcon(), func() {
		c.validateCanvas()
	})

	historyAction := widget.NewToolbarAction(theme.HistoryIcon(), func() {
		c.showVersionHistory()
	})

	settingsAction := widget.NewToolbarAction(theme.SettingsIcon(), func() {
		c.showSettings()
	})

	return widget.NewToolbar(
		saveAction,
		loadAction,
		widget.NewToolbarSeparator(),
		exportAction,
		validateAction,
		widget.NewToolbarSeparator(),
		historyAction,
		settingsAction,
		widget.NewToolbarSeparator(),
		themeToggle,
	)
}

func (c *Canvas) createMainContent() *fyne.Container {
	// Create section containers with tooltips
	keyPartnersContainer := createSection("Key Partners", c.keyPartners, "Who are your key partners and suppliers? What resources are you acquiring from them?")
	keyActivitiesContainer := createSection("Key Activities", c.keyActivities, "What key activities does your value proposition require?")
	keyResourcesContainer := createSection("Key Resources", c.keyResources, "What key resources does your value proposition require?")
	valuePropContainer := createSection("Value Proposition", c.valueProposition, "What value do you deliver to customers? Which problems are you solving?")
	customerRelContainer := createSection("Customer Relationships", c.customerRel, "What type of relationship does each customer segment expect?")
	channelsContainer := createSection("Channels", c.channels, "Through which channels do your customers want to be reached?")
	customerSegContainer := createSection("Customer Segments", c.customerSegments, "For whom are you creating value? Who are your most important customers?")
	costContainer := createSection("Cost Structure", c.costStructure, "What are the most important costs inherent in your business model?")
	revenueContainer := createSection("Revenue Streams", c.revenueStreams, "For what value are your customers willing to pay? How would they prefer to pay?")

	// Create the top grid
	topGrid := container.NewGridWithColumns(5,
		keyPartnersContainer,
		container.NewGridWithRows(2,
			keyActivitiesContainer,
			keyResourcesContainer,
		),
		valuePropContainer,
		container.NewGridWithRows(2,
			customerRelContainer,
			channelsContainer,
		),
		customerSegContainer,
	)

	// Create the bottom grid
	bottomGrid := container.NewGridWithColumns(2,
		costContainer,
		revenueContainer,
	)

	// Combine top and bottom grids
	return container.NewGridWithRows(2,
		topGrid,
		bottomGrid,
	)
}

// HoverableRect implements desktop.Hoverable
type HoverableRect struct {
	canvas.Rectangle
	popup   *widget.PopUp
	tooltip string
}

func NewHoverableRect(tooltip string) *HoverableRect {
	rect := &HoverableRect{
		tooltip: tooltip,
	}
	rect.FillColor = color.Transparent
	return rect
}

func (h *HoverableRect) MouseIn(e *desktop.MouseEvent) {
	if h.popup == nil {
		tooltipLabel := widget.NewLabel(h.tooltip)
		canvas := fyne.CurrentApp().Driver().AllWindows()[0].Canvas()
		h.popup = widget.NewPopUp(tooltipLabel, canvas)
	}
	h.popup.ShowAtPosition(e.Position)
}

func (h *HoverableRect) MouseOut() {
	if h.popup != nil {
		h.popup.Hide()
	}
}

func (h *HoverableRect) MouseMoved(*desktop.MouseEvent) {
}

func createSection(title string, entry *widget.Entry, tooltip string) *fyne.Container {
	label := widget.NewLabel(title)

	// Create a container for the entry
	entryContainer := container.NewStack(entry)

	// Add hoverable area
	hoverArea := NewHoverableRect(tooltip)
	hoverArea.Resize(entry.Size())
	entryContainer.Add(hoverArea)

	return container.NewBorder(
		label, nil, nil, nil,
		container.NewPadded(entryContainer),
	)
}

func (c *Canvas) createStatusBar() *fyne.Container {
	return container.NewHBox(
		widget.NewLabel("Status: Ready"),
		c.progressBar,
	)
}

func (c *Canvas) showSettings() {
	// Create settings form
	autoSaveCheck := widget.NewCheck("Auto-save", func(checked bool) {
		c.autoSave = checked
	})
	autoSaveCheck.SetChecked(c.autoSave)

	currentThemeLabel := widget.NewLabel("Current Theme: " + c.currentTheme)

	type ThemeOption struct {
		Name  string
		Value string
	}

	var themeOptions = []ThemeOption{
		{Name: "Professional (Dark)", Value: "professional"},
		{Name: "Light", Value: "light"},
	}

	themeSelect := widget.NewSelect([]string{"Professional (Dark)", "Light"}, func(selected string) {
		for _, option := range themeOptions {
			if option.Name == selected {
				c.currentTheme = option.Value
				myApp := fyne.CurrentApp()
				if c.currentTheme == "professional" {
					myApp.Settings().SetTheme(theme.DarkTheme())
				} else {
					myApp.Settings().SetTheme(theme.LightTheme())
				}
				break
			}
		}
	})

	themeSelect.SetSelected("Professional (Dark)")

	checkFormItem := widget.NewFormItem("Auto-save", autoSaveCheck)
	themeFormItem := widget.NewFormItem("Theme", themeSelect)

	itemList := []*widget.FormItem{checkFormItem, themeFormItem}

	if c.currentTheme == "professional" {
		currentThemeLabel.SetText("Current Theme: Professional (Dark)")
	} else {
		currentThemeLabel.SetText("Current Theme: Light")
	}

	infoContainer := container.NewVBox(currentThemeLabel)
	infoContainer.Add(widget.NewLabel("Change settings below:"))

	infoAndForm := container.NewBorder(infoContainer, nil, nil, nil, &widget.Form{Items: itemList})

	dialog.ShowCustom("Settings", "Close", infoAndForm, c.window)
}

func (c *Canvas) autoSaveRoutine() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		if c.autoSave && time.Since(c.lastSaved) >= 5*time.Minute {
			c.saveCurrentVersion()
		}
	}
}

func (c *Canvas) saveCurrentVersion() {
	version := Version{
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		Data:      c.getCurrentData(),
	}
	c.versions = append(c.versions, version)
	c.lastSaved = time.Now()

	// Update progress
	c.updateProgress()
}

func (c *Canvas) getCurrentData() CanvasData {
	return CanvasData{
		KeyPartners:      c.keyPartners.Text,
		KeyActivities:    c.keyActivities.Text,
		KeyResources:     c.keyResources.Text,
		ValueProposition: c.valueProposition.Text,
		CustomerRel:      c.customerRel.Text,
		Channels:         c.channels.Text,
		CustomerSegments: c.customerSegments.Text,
		CostStructure:    c.costStructure.Text,
		RevenueStreams:   c.revenueStreams.Text,
	}
}

func (c *Canvas) updateProgress() {
	// Calculate progress based on filled sections
	totalSections := 9.0
	filledSections := 0.0

	if len(c.keyPartners.Text) > 0 {
		filledSections++
		c.updateSectionColor(c.keyPartners, true)
	} else {
		c.updateSectionColor(c.keyPartners, false)
	}
	if len(c.keyActivities.Text) > 0 {
		filledSections++
		c.updateSectionColor(c.keyActivities, true)
	} else {
		c.updateSectionColor(c.keyActivities, false)
	}
	if len(c.keyResources.Text) > 0 {
		filledSections++
		c.updateSectionColor(c.keyResources, true)
	} else {
		c.updateSectionColor(c.keyResources, false)
	}
	if len(c.valueProposition.Text) > 0 {
		filledSections++
		c.updateSectionColor(c.valueProposition, true)
	} else {
		c.updateSectionColor(c.valueProposition, false)
	}
	if len(c.customerRel.Text) > 0 {
		filledSections++
		c.updateSectionColor(c.customerRel, true)
	} else {
		c.updateSectionColor(c.customerRel, false)
	}
	if len(c.channels.Text) > 0 {
		filledSections++
		c.updateSectionColor(c.channels, true)
	} else {
		c.updateSectionColor(c.channels, false)
	}
	if len(c.customerSegments.Text) > 0 {
		filledSections++
		c.updateSectionColor(c.customerSegments, true)
	} else {
		c.updateSectionColor(c.customerSegments, false)
	}
	if len(c.costStructure.Text) > 0 {
		filledSections++
		c.updateSectionColor(c.costStructure, true)
	} else {
		c.updateSectionColor(c.costStructure, false)
	}
	if len(c.revenueStreams.Text) > 0 {
		filledSections++
		c.updateSectionColor(c.revenueStreams, true)
	} else {
		c.updateSectionColor(c.revenueStreams, false)
	}

	c.progressBar.SetValue(filledSections / totalSections)
}

func (c *Canvas) validateCanvas() {
	results := c.validator.Validate(c)
	if len(results) > 0 {
		var message string
		for _, result := range results {
			message += fmt.Sprintf("• %s: %s\n", result.Section, result.Message)
		}
		dialog.ShowInformation("Validation Results", message, c.window)
	} else {
		dialog.ShowInformation("Validation Results", "All sections look good!", c.window)
	}
}

func (c *Canvas) showVersionHistory() {
	if len(c.versions) == 0 {
		dialog.ShowInformation("Version History", "No previous versions found", c.window)
		return
	}

	var items []string
	for _, version := range c.versions {
		items = append(items, version.Timestamp.Format("2006-01-02 15:04:05"))
	}

	list := widget.NewList(
		func() int { return len(items) },
		func() fyne.CanvasObject { return widget.NewLabel("Template") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(items[id])
		},
	)

	// Add restore button
	list.OnSelected = func(id widget.ListItemID) {
		dialog.ShowConfirm("Restore Version", "Do you want to restore this version?",
			func(restore bool) {
				if restore {
					c.restoreVersion(c.versions[id])
				}
			}, c.window)
	}

	dialog.ShowCustom("Version History", "Close", list, c.window)
}

func (c *Canvas) restoreVersion(version Version) {
	// Save current state to undo stack
	c.undoStack = append(c.undoStack, c.getCurrentData())

	// Restore the selected version
	c.keyPartners.SetText(version.Data.KeyPartners)
	c.keyActivities.SetText(version.Data.KeyActivities)
	c.keyResources.SetText(version.Data.KeyResources)
	c.valueProposition.SetText(version.Data.ValueProposition)
	c.customerRel.SetText(version.Data.CustomerRel)
	c.channels.SetText(version.Data.Channels)
	c.customerSegments.SetText(version.Data.CustomerSegments)
	c.costStructure.SetText(version.Data.CostStructure)
	c.revenueStreams.SetText(version.Data.RevenueStreams)

	// Update progress and colors
	c.updateProgress()

	dialog.ShowInformation("Success", "Version restored successfully", c.window)
}

// BusinessValidator handles canvas validation
type BusinessValidator struct {
	rules []ValidationRule
}

type ValidationRule struct {
	Section string
	Check   func(*Canvas) bool
	Message string
}

type ValidationResult struct {
	Section string
	Message string
}

func NewBusinessValidator() *BusinessValidator {
	return &BusinessValidator{
		rules: []ValidationRule{
			{
				Section: "Value Proposition",
				Check: func(c *Canvas) bool {
					return len(c.valueProposition.Text) >= 100
				},
				Message: "Consider adding more detail about your value proposition",
			},
			{
				Section: "Customer Segments",
				Check: func(c *Canvas) bool {
					return len(c.customerSegments.Text) >= 50
				},
				Message: "Customer segments need more specific details",
			},
			{
				Section: "Key Activities",
				Check: func(c *Canvas) bool {
					return len(c.keyActivities.Text) >= 50
				},
				Message: "Add more details about your key activities",
			},
			{
				Section: "Cost Structure",
				Check: func(c *Canvas) bool {
					return len(c.costStructure.Text) >= 50
				},
				Message: "Elaborate on your cost structure",
			},
			{
				Section: "Revenue Streams",
				Check: func(c *Canvas) bool {
					return len(c.revenueStreams.Text) >= 50
				},
				Message: "Provide more information about revenue streams",
			},
		},
	}
}

func (v *BusinessValidator) Validate(canvas *Canvas) []ValidationResult {
	var results []ValidationResult

	for _, rule := range v.rules {
		if !rule.Check(canvas) {
			results = append(results, ValidationResult{
				Section: rule.Section,
				Message: rule.Message,
			})
		}
	}

	return results
}

func (c *Canvas) setupKeyboardShortcuts() {
	c.window.Canvas().AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyS, Modifier: fyne.KeyModifierControl},
		func(shortcut fyne.Shortcut) {
			c.saveCanvas()
		},
	)

	c.window.Canvas().AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyO, Modifier: fyne.KeyModifierControl},
		func(shortcut fyne.Shortcut) {
			c.loadCanvas()
		},
	)

	c.window.Canvas().AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyZ, Modifier: fyne.KeyModifierControl},
		func(shortcut fyne.Shortcut) {
			c.undo()
		},
	)

	c.window.Canvas().AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyY, Modifier: fyne.KeyModifierControl},
		func(shortcut fyne.Shortcut) {
			c.redo()
		},
	)

	c.window.Canvas().AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyP, Modifier: fyne.KeyModifierControl},
		func(shortcut fyne.Shortcut) {
			c.exportToPDF()
		},
	)

	// For clipboard operations
	c.window.Canvas().AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyC, Modifier: fyne.KeyModifierControl},
		func(shortcut fyne.Shortcut) {
			focusedEntry := c.window.Canvas().Focused()
			if entry, ok := focusedEntry.(*widget.Entry); ok {
				c.window.Clipboard().SetContent(entry.Text)
			} else {
				fyne.LogError("Clipboard copy failed", errors.New("no focused entry or invalid type"))
			}
		},
	)

	c.window.Canvas().AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyV, Modifier: fyne.KeyModifierControl},
		func(shortcut fyne.Shortcut) {
			focusedEntry := c.window.Canvas().Focused()
			if entry, ok := focusedEntry.(*widget.Entry); ok {
				entry.SetText(c.window.Clipboard().Content())
			} else {
				fyne.LogError("Clipboard paste failed", errors.New("no focused entry or invalid type"))
			}
		},
	)

	c.window.Canvas().AddShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyX, Modifier: fyne.KeyModifierControl},
		func(shortcut fyne.Shortcut) {
			focusedEntry := c.window.Canvas().Focused()
			if entry, ok := focusedEntry.(*widget.Entry); ok {
				c.window.Clipboard().SetContent(entry.Text)
				entry.SetText("")
			} else {
				fyne.LogError("Clipboard cut failed", errors.New("no focused entry or invalid type"))
			}
		},
	)
}

func (c *Canvas) setupDynamicValidation(entry *widget.Entry, section string) {
	entry.OnChanged = func(s string) {
		results := c.validator.Validate(c)
		isValid := true
		for _, result := range results {
			if result.Section == section {
				isValid = false
				break
			}
		}
		c.updateSectionColor(entry, isValid)
	}
}

func (c *Canvas) updateSectionColor(entry *widget.Entry, isValid bool) {
	if isValid {
		entry.TextStyle = fyne.TextStyle{} // Reset to default
	} else {
		entry.TextStyle = fyne.TextStyle{Italic: true} // Visual indicator for invalid
	}
	entry.Refresh() // Refresh the widget to show changes
}

func (c *Canvas) undo() {
	if len(c.undoStack) > 0 {
		// Save current state to redo stack
		c.redoStack = append(c.redoStack, c.getCurrentData())

		// Pop the last state from undo stack
		lastState := c.undoStack[len(c.undoStack)-1]
		c.undoStack = c.undoStack[:len(c.undoStack)-1]

		// Restore the state
		c.keyPartners.SetText(lastState.KeyPartners)
		c.keyActivities.SetText(lastState.KeyActivities)
		c.keyResources.SetText(lastState.KeyResources)
		c.valueProposition.SetText(lastState.ValueProposition)
		c.customerRel.SetText(lastState.CustomerRel)
		c.channels.SetText(lastState.Channels)
		c.customerSegments.SetText(lastState.CustomerSegments)
		c.costStructure.SetText(lastState.CostStructure)
		c.revenueStreams.SetText(lastState.RevenueStreams)

		// Update progress and colors
		c.updateProgress()
	}
}

func (c *Canvas) redo() {
	if len(c.redoStack) > 0 {
		// Save current state to undo stack
		c.undoStack = append(c.undoStack, c.getCurrentData())

		// Pop the last state from redo stack
		lastState := c.redoStack[len(c.redoStack)-1]
		c.redoStack = c.redoStack[:len(c.redoStack)-1]

		// Restore the state
		c.keyPartners.SetText(lastState.KeyPartners)
		c.keyActivities.SetText(lastState.KeyActivities)
		c.keyResources.SetText(lastState.KeyResources)
		c.valueProposition.SetText(lastState.ValueProposition)
		c.customerRel.SetText(lastState.CustomerRel)
		c.channels.SetText(lastState.Channels)
		c.customerSegments.SetText(lastState.CustomerSegments)
		c.costStructure.SetText(lastState.CostStructure)
		c.revenueStreams.SetText(lastState.RevenueStreams)

		// Update progress and colors
		c.updateProgress()
	}
}

func (c *Canvas) saveCanvas() {
	dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, c.window)
			return
		}
		if writer == nil {
			return
		}
		defer writer.Close()

		// Save current state to undo stack
		c.undoStack = append(c.undoStack, c.getCurrentData())

		// Prepare data
		data := c.getCurrentData()
		jsonData, err := json.MarshalIndent(data, "", "    ")
		if err != nil {
			dialog.ShowError(err, c.window)
			return
		}

		// Write to file
		_, err = writer.Write(jsonData)
		if err != nil {
			dialog.ShowError(err, c.window)
			return
		}

		dialog.ShowInformation("Success", "Canvas saved successfully", c.window)
	}, c.window)
}

func (c *Canvas) loadCanvas() {
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, c.window)
			return
		}
		if reader == nil {
			return
		}
		defer reader.Close()

		// Save current state to undo stack
		c.undoStack = append(c.undoStack, c.getCurrentData())

		// Read file contents
		data, err := io.ReadAll(reader)
		if err != nil {
			dialog.ShowError(err, c.window)
			return
		}

		// Parse JSON
		var canvasData CanvasData
		err = json.Unmarshal(data, &canvasData)
		if err != nil {
			dialog.ShowError(err, c.window)
			return
		}

		// Update canvas fields
		c.keyPartners.SetText(canvasData.KeyPartners)
		c.keyActivities.SetText(canvasData.KeyActivities)
		c.keyResources.SetText(canvasData.KeyResources)
		c.valueProposition.SetText(canvasData.ValueProposition)
		c.customerRel.SetText(canvasData.CustomerRel)
		c.channels.SetText(canvasData.Channels)
		c.customerSegments.SetText(canvasData.CustomerSegments)
		c.costStructure.SetText(canvasData.CostStructure)
		c.revenueStreams.SetText(canvasData.RevenueStreams)

		// Update progress and colors
		c.updateProgress()

		dialog.ShowInformation("Success", "Canvas loaded successfully", c.window)
	}, c.window)
}

func (c *Canvas) exportToPDF() {
	pdf := gofpdf.New("L", "mm", "A3", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)

	// Page settings
	pageWidth := 420.0  // A3 landscape width
	pageHeight := 297.0 // A3 landscape height
	margin := 10.0

	// Calculate section dimensions
	topHeight := (pageHeight - 2*margin) * 0.6
	bottomHeight := (pageHeight - 2*margin) * 0.4
	colWidth := (pageWidth - 2*margin) / 5

	// Draw borders and titles
	pdf.SetLineWidth(0.3)

	// Top sections
	y := margin
	// Key Partners
	drawSection(pdf, margin, y, colWidth, topHeight, "Key Partners", c.keyPartners.Text)

	// Key Activities & Resources
	x := margin + colWidth
	drawSection(pdf, x, y, colWidth, topHeight/2, "Key Activities", c.keyActivities.Text)
	drawSection(pdf, x, y+topHeight/2, colWidth, topHeight/2, "Key Resources", c.keyResources.Text)

	// Value Proposition
	x += colWidth
	drawSection(pdf, x, y, colWidth, topHeight, "Value Proposition", c.valueProposition.Text)

	// Customer Relationships & Channels
	x += colWidth
	drawSection(pdf, x, y, colWidth, topHeight/2, "Customer Relationships", c.customerRel.Text)
	drawSection(pdf, x, y+topHeight/2, colWidth, topHeight/2, "Channels", c.channels.Text)

	// Customer Segments
	x += colWidth
	drawSection(pdf, x, y, colWidth, topHeight, "Customer Segments", c.customerSegments.Text)

	// Bottom sections
	y = margin + topHeight
	// Cost Structure
	drawSection(pdf, margin, y, (pageWidth-2*margin)/2, bottomHeight, "Cost Structure", c.costStructure.Text)

	// Revenue Streams
	drawSection(pdf, margin+(pageWidth-2*margin)/2, y, (pageWidth-2*margin)/2, bottomHeight, "Revenue Streams", c.revenueStreams.Text)

	// Save PDF
	dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, c.window)
			return
		}
		if writer == nil {
			return
		}
		defer writer.Close()

		err = pdf.Output(writer)
		if err != nil {
			dialog.ShowError(err, c.window)
			return
		}

		dialog.ShowInformation("Success", "PDF has been exported successfully", c.window)
	}, c.window)
}

func drawSection(pdf *gofpdf.Fpdf, x, y, w, h float64, title, content string) {
	pdf.Rect(x, y, w, h, "D") // "D" means draw border only

	// Draw title
	pdf.SetFont("Arial", "B", 12)
	pdf.Text(x+5, y+10, title)

	// Draw content
	pdf.SetFont("Arial", "", 10)
	pdf.SetXY(x+5, y+15)
	pdf.MultiCell(w-10, 5, content, "", "", false)
}
