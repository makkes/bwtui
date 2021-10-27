// SPDX-License-Identifier: GPL-3.0-or-later
// (c) 2021 Max Jonas Werner <mail@makk.es>

package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Login struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Object struct {
	Type  string `json:"object"`
	Name  string `json:"name"`
	Login *Login `json:"login"`
}

func getItems(s string) []Object {
	args := []string{"list", "items"}
	if s != "" {
		args = append(args, "--search", s)
	}
	cmd := exec.Command("bw", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Run()
	var obj []Object
	if err := json.Unmarshal(out.Bytes(), &obj); err != nil {
		panic(err)
	}
	return obj
}

func copyToClipboard(s string) error {
	cmd := exec.Command("xsel", "-b")
	cmd.Stdin = strings.NewReader(s)
	return cmd.Run()
}

func main() {
	search := ""
	if len(os.Args) >= 2 {
		search = os.Args[1]
	}
	app := tview.NewApplication()
	list := tview.NewList()
	bgColor := list.GetBackgroundColor()
	fgColor := list.GetBorderColor()
	list.SetBackgroundColor(fgColor)
	list.SetBorderColor(bgColor)
	list.SetMainTextColor(bgColor)
	list.SetSelectedBackgroundColor(bgColor)
	list.SetSelectedTextColor(fgColor)
	items := getItems(search)
	for _, item := range items {
		list.AddItem(item.Name, "", 0, nil)
	}
	list.ShowSecondaryText(false)
	list.SetHighlightFullLine(true)
	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune {
			if event.Rune() == 'p' {
				copyToClipboard(items[list.GetCurrentItem()].Login.Password)
			}
			if event.Rune() == 'u' {
				copyToClipboard(items[list.GetCurrentItem()].Login.Username)
			}
			if event.Rune() == 'q' {
				app.Stop()
			}
		}
		return event
	})

	if err := app.SetRoot(list, true).Run(); err != nil {
		panic(err)
	}
}
