package diff

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
	"golang.org/x/crypto/ssh/terminal"
)

const (
	green  = "\u001b[32m"
	yellow = "\u001b[33m"
	red    = "\u001b[31m"
	reset  = "\u001b[0m"
)

func (i Item) Pretty() string {
	if i.Key == "" {
		return ""
	}
	return fmt.Sprintf("%s: %v", i.Key, i.Value)
}

func (d ChangesPresentDiff) Pretty(colorEnabled bool) string {
	var padding int

	terminalWidth, _, err := terminal.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		log.Println("Couldn't get terminal size: " + err.Error())
		terminalWidth = 120
	}

	for _, delta := range d.Deltas() {
		l := len(delta.Key())
		if l < terminalWidth/2 && l > padding {
			padding = l
		}
	}

	printer := colorPrinter{colorEnabled: colorEnabled}

	prettyStr := bytes.NewBuffer(nil)
	for _, delta := range d.Deltas() {
		sourceVal := strOrRepr(delta.SourceItem.Value)
		serverVal := strOrRepr(delta.ServerItem.Value)
		if multilineString(sourceVal) || multilineString(serverVal) {
			dmp := diffmatchpatch.New()
			diffs := dmp.DiffMain(serverVal, sourceVal, false)
			prettyStr.WriteString(dmp.DiffPrettyText(diffs))
		} else {
			prettyStr.WriteString(delta.DiffString(printer, padding))
		}
		prettyStr.WriteRune('\n')
	}
	if colorEnabled {
		prettyStr.WriteString(reset)
	}
	return prettyStr.String()
}

func strOrRepr(v interface{}) string {
	s, ok := v.(string)
	if !ok {
		return fmt.Sprintf("%v", v)
	}
	return s
}

func multilineString(s string) bool {
	return strings.Index(s, "\n") != -1
}

func (d Delta) DiffString(printer colorPrinter, padding int) string {
	if (d.SourceItem != Item{} && d.ServerItem == Item{}) {
		return printer.Print(green, fmt.Sprintf("+ %-*s: %q", padding, d.Key(), d.SourceItem.Value))
	} else if (d.SourceItem != Item{} && d.ServerItem != Item{}) {
		return printer.Print(
			yellow,
			fmt.Sprintf("~ %-*s: %s => %s",
				padding, d.Key(),
				printer.Print(red, fmt.Sprintf("%q", d.ServerItem.Value)),
				printer.Print(green, fmt.Sprintf("%q", d.SourceItem.Value)),
			),
		)
	} else if (d.SourceItem == Item{} && d.ServerItem != Item{}) {
		return printer.Print(red, fmt.Sprintf("- %-*s: %q", padding, d.Key(), d.ServerItem.Value))
	} else {
		panic("comparing two empty items, this should not happen")
	}
}

type colorPrinter struct {
	colorEnabled bool
}

func (c colorPrinter) Print(color string, str string) string {
	if !c.colorEnabled {
		return str
	}
	return fmt.Sprintf("%s%s%s", color, str, reset)
}
