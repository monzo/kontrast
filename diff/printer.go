package diff

import (
	"fmt"
	"os"

	"golang.org/x/crypto/ssh/terminal"
)

const (
	green = "\u001b[32m"
	cyan  = "\u001b[36m"
	red   = "\u001b[31m"
	reset = "\u001b[0m"
)

func (i Item) Pretty() string {
	if i.Key == "" {
		return ""
	}
	return fmt.Sprintf("%s: %v", i.Key, i.Value)
}

func (d ChangesPresentDiff) Pretty() string {
	var maxLeft, maxRight int

	terminalWidth, _, err := terminal.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		panic("Couldn't get terminal size: " + err.Error())
	}

	for _, delta := range d.Deltas() {
		l := len(delta.SourceItem.Pretty())
		if l < terminalWidth/2 && l > maxLeft {
			maxLeft = l
		}

		r := len(delta.ServerItem.Pretty())
		if r < terminalWidth/2 && r > maxRight {
			maxRight = r
		}
	}

	prettyStr := ""
	for _, delta := range d.Deltas() {
		prettyStr += fmt.Sprintf(
			"%s %-*s | %-*s%s\n",
			delta.getPrettyLineStart(),
			maxLeft, delta.SourceItem.Pretty(),
			maxRight, delta.ServerItem.Pretty(),
			reset,
		)
	}
	return prettyStr
}

func (d Delta) getPrettyLineStart() string {
	gutterChar := ""
	lineColor := ""
	if (d.SourceItem != Item{} && d.ServerItem == Item{}) {
		gutterChar = "+"
		lineColor = green
	} else if (d.SourceItem != Item{} && d.ServerItem != Item{}) {
		gutterChar = "~"
		lineColor = cyan
	} else if (d.SourceItem == Item{} && d.ServerItem != Item{}) {
		gutterChar = "-"
		lineColor = red
	}
	return fmt.Sprintf("%s%s", lineColor, gutterChar)
}
