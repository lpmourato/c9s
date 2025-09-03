package tui

import (
	"github.com/derailed/tcell/v2"
)

// KeyAction represents a keyboard action
type KeyAction int

const (
	ActionShowLogs KeyAction = iota
	ActionShowDeploymentDetails
	ActionShowServiceDescription
	ActionToggleCommand
	ActionEscape
	ActionQuit
	ActionEnter
	ActionScrollUp
	ActionScrollDown
	ActionFilter
	ActionRefresh
	ActionHelp
)

// KeyHandler represents a centralized keyboard input handler
type KeyHandler struct {
	app          *App
	bindings     map[tcell.Key]KeyAction
	runeBindings map[rune]KeyAction
	handlers     map[KeyAction]func() error
}

// NewKeyHandler creates a new centralized key handler
func NewKeyHandler(app *App) *KeyHandler {
	kh := &KeyHandler{
		app:          app,
		bindings:     make(map[tcell.Key]KeyAction),
		runeBindings: make(map[rune]KeyAction),
		handlers:     make(map[KeyAction]func() error),
	}

	// Set up default key bindings
	kh.setupDefaultBindings()

	return kh
}

// setupDefaultBindings configures the default keyboard shortcuts
func (kh *KeyHandler) setupDefaultBindings() {
	// Special keys
	kh.bindings[tcell.KeyEscape] = ActionEscape
	kh.bindings[tcell.KeyEnter] = ActionEnter
	kh.bindings[tcell.KeyUp] = ActionScrollUp
	kh.bindings[tcell.KeyDown] = ActionScrollDown
	kh.bindings[tcell.KeyCtrlC] = ActionQuit

	// Character keys
	kh.runeBindings[':'] = ActionToggleCommand
	kh.runeBindings['l'] = ActionShowLogs
	kh.runeBindings['L'] = ActionShowLogs
	kh.runeBindings['d'] = ActionShowDeploymentDetails
	kh.runeBindings['D'] = ActionShowDeploymentDetails
	kh.runeBindings['s'] = ActionShowServiceDescription
	kh.runeBindings['S'] = ActionShowServiceDescription
	kh.runeBindings['q'] = ActionQuit
	kh.runeBindings['Q'] = ActionQuit
	kh.runeBindings['r'] = ActionRefresh
	kh.runeBindings['R'] = ActionRefresh
	kh.runeBindings['f'] = ActionFilter
	kh.runeBindings['F'] = ActionFilter
	kh.runeBindings['h'] = ActionHelp
	kh.runeBindings['H'] = ActionHelp
	kh.runeBindings['?'] = ActionHelp
}

// RegisterHandler registers a handler function for a specific action
func (kh *KeyHandler) RegisterHandler(action KeyAction, handler func() error) {
	kh.handlers[action] = handler
}

// RegisterKeyBinding registers a custom key binding
func (kh *KeyHandler) RegisterKeyBinding(key tcell.Key, action KeyAction) {
	kh.bindings[key] = action
}

// RegisterRuneBinding registers a custom rune binding
func (kh *KeyHandler) RegisterRuneBinding(r rune, action KeyAction) {
	kh.runeBindings[r] = action
}

// HandleEvent processes a keyboard event and executes the appropriate action
func (kh *KeyHandler) HandleEvent(event *tcell.EventKey) *tcell.EventKey {
	var action KeyAction
	var found bool

	// Check for special key bindings first
	if event.Key() != tcell.KeyRune {
		action, found = kh.bindings[event.Key()]
	} else {
		// Check for rune bindings
		action, found = kh.runeBindings[event.Rune()]
	}

	if found {
		if handler, exists := kh.handlers[action]; exists {
			if err := handler(); err != nil {
				// Log error or handle it appropriately
				// For now, we'll just continue
			}
			return nil // Event consumed
		}
	}

	// Event not handled, pass it through
	return event
}

// CreateInputCapture creates an input capture function for a tview component
func (kh *KeyHandler) CreateInputCapture() func(*tcell.EventKey) *tcell.EventKey {
	return kh.HandleEvent
}

// CreateInputCaptureWithFilter creates an input capture with a custom filter
// The filter function should return true if the event should be processed by KeyHandler
func (kh *KeyHandler) CreateInputCaptureWithFilter(filter func(*tcell.EventKey) bool) func(*tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		if filter(event) {
			return kh.HandleEvent(event)
		}
		return event
	}
}

// GetAvailableKeys returns a list of available key bindings for help display
func (kh *KeyHandler) GetAvailableKeys() map[string]string {
	keys := make(map[string]string)

	// Add descriptions for common keys
	keys["Enter"] = "Show logs for selected service"
	keys["d/D"] = "Show deployment details"
	keys["s/S"] = "Show service description"
	keys[":"] = "Open command input"
	keys["l/L"] = "Show logs"
	keys["r/R"] = "Refresh data"
	keys["f/F"] = "Filter services"
	keys["h/H/?"] = "Show help"
	keys["q/Q"] = "Quit application"
	keys["Esc"] = "Go back/Cancel"
	keys["↑/↓"] = "Scroll up/down"
	keys["Ctrl+C"] = "Force quit"

	return keys
}

// Context represents the current UI context for conditional key handling
type Context int

const (
	ContextMain Context = iota
	ContextCommandInput
	ContextLogView
	ContextDeploymentView
	ContextServiceDetails
)

// ContextualKeyHandler extends KeyHandler with context awareness
type ContextualKeyHandler struct {
	*KeyHandler
	currentContext Context
	contextFilters map[Context]func(*tcell.EventKey) bool
}

// NewContextualKeyHandler creates a new context-aware key handler
func NewContextualKeyHandler(app *App) *ContextualKeyHandler {
	ckh := &ContextualKeyHandler{
		KeyHandler:     NewKeyHandler(app),
		currentContext: ContextMain,
		contextFilters: make(map[Context]func(*tcell.EventKey) bool),
	}

	// Set up context filters
	ckh.setupContextFilters()

	return ckh
}

// setupContextFilters configures filters for different contexts
func (ckh *ContextualKeyHandler) setupContextFilters() {
	// Main context - handle most keys
	ckh.contextFilters[ContextMain] = func(event *tcell.EventKey) bool {
		return true // Handle all keys in main context
	}

	// Command input context - only handle escape
	ckh.contextFilters[ContextCommandInput] = func(event *tcell.EventKey) bool {
		return event.Key() == tcell.KeyEscape
	}

	// Log view context - handle navigation and escape
	ckh.contextFilters[ContextLogView] = func(event *tcell.EventKey) bool {
		return event.Key() == tcell.KeyEscape ||
			event.Key() == tcell.KeyUp ||
			event.Key() == tcell.KeyDown ||
			(event.Key() == tcell.KeyRune && (event.Rune() == 'q' || event.Rune() == 'Q'))
	}

	// Deployment view context - similar to log view
	ckh.contextFilters[ContextDeploymentView] = func(event *tcell.EventKey) bool {
		return event.Key() == tcell.KeyEscape ||
			event.Key() == tcell.KeyUp ||
			event.Key() == tcell.KeyDown ||
			(event.Key() == tcell.KeyRune && (event.Rune() == 'q' || event.Rune() == 'Q'))
	}
}

// SetContext changes the current context
func (ckh *ContextualKeyHandler) SetContext(context Context) {
	ckh.currentContext = context
}

// GetContext returns the current context
func (ckh *ContextualKeyHandler) GetContext() Context {
	return ckh.currentContext
}

// CreateContextualInputCapture creates an input capture that respects the current context
func (ckh *ContextualKeyHandler) CreateContextualInputCapture() func(*tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		if filter, exists := ckh.contextFilters[ckh.currentContext]; exists {
			if filter(event) {
				return ckh.HandleEvent(event)
			}
		}
		return event
	}
}
