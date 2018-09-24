package ui

// UI is an interface for implemenations of interactivity
type UI interface {
	Prompt(string, string, string) bool
}
