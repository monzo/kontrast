package diff

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
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

func (d ChangesPresentDiff) Pretty(colorEnabled bool) string {
	var maxLeft, maxRight int

	terminalWidth, _, err := terminal.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		log.Println("Couldn't get terminal size: " + err.Error())
		terminalWidth = 120
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
		sourceVal := strOrRepr(delta.SourceItem.Value)
		serverVal := strOrRepr(delta.ServerItem.Value)
		if multilineString(sourceVal) || multilineString(serverVal) {
			dmp := diffmatchpatch.New()
			diffs := dmp.DiffMain(sourceVal, serverVal, false)
			prettyStr += dmp.DiffPrettyText(diffs)
		} else {
			prettyStr += fmt.Sprintf(
				"%s %-*s | %-*s",
				delta.getPrettyLineStart(colorEnabled),
				maxLeft, delta.SourceItem.Pretty(),
				maxRight, delta.ServerItem.Pretty(),
			)
		}
	}
	if colorEnabled {
		prettyStr += reset
	}
	return prettyStr
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

func (d Delta) getPrettyLineStart(colorEnabled bool) string {
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
	if !colorEnabled {
		lineColor = ""
	}
	return fmt.Sprintf("%s%s", lineColor, gutterChar)
}
