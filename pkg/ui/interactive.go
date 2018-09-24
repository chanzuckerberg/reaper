package ui

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	input "github.com/tcnksm/go-input"
)

// Interactive will prompt a user
type Interactive struct {
	prompt *input.UI
}

func NewInteractive() *Interactive {
	prompt := &input.UI{
		Writer: os.Stdout,
		Reader: os.Stdin,
	}
	return &Interactive{prompt: prompt}
}

func (i *Interactive) Prompt(msg, recipient, method string) bool {
	log.Info(msg)
	a := fmt.Sprintf("---------------\nI want to send:\n\n\t%s\n\nto: %s\nvia %s.\n\nShould I?", msg, recipient, method)
	yes, err := i.prompt.Ask(a, &input.Options{
		Required: true,
		Default:  "Y",
	})

	if err != nil {
		log.Debugf("error: %#v", err)
		return false
	}
	return yes == "Y" || yes == "y"
}
