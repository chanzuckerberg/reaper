package ui

import (
	"bytes"
	"os"
	"text/tabwriter"
	"text/template"

	"github.com/Masterminds/sprig"
	log "github.com/sirupsen/logrus"
	input "github.com/tcnksm/go-input"
)

// Interactive will prompt a user
type Interactive struct {
	prompt *input.UI
}

//NewInteractive will create a interactive ui object
func NewInteractive() *Interactive {
	prompt := &input.UI{
		Writer: os.Stdout,
		Reader: os.Stdin,
	}
	return &Interactive{prompt: prompt}
}

// Prompt will give the user `msg` and prompt for confirmation
func (i *Interactive) Prompt(msg, recipient, method string) bool {
	log.Info(msg)

	data := map[string]string{
		"msg":       msg,
		"recipient": recipient,
		"method":    method,
	}

	tmpl := "---------------\nI want to send:\n\n{{.msg | indent 4}}\n\nto: {{.recipient}}\nvia {{.method}}.\n\nShould I?"

	t, err := template.New("message").Funcs(sprig.TxtFuncMap()).Parse(tmpl)
	if err != nil {
		// FIXME
		panic(err)
	}

	messageBytes := bytes.NewBuffer(nil)
	// tabwriter will take consectutive lines with tabs in them
	// and do a column alignment
	w := tabwriter.NewWriter(messageBytes, 4, 4, 4, ' ', 0)
	err = t.Execute(w, data)
	if err != nil {
		panic("Could not template message")
	}

	message := messageBytes.String()

	yes, err := i.prompt.Ask(message, &input.Options{
		Required: true,
		Default:  "Y",
	})

	if err != nil {
		if err.Error() == "interrupted" {
			// FIXME this is def not what we really want to do
			os.Exit(-1)
		}
		log.Infof("error: %#v", err)
		return false
	}
	return yes == "Y" || yes == "y"
}
