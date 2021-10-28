// SPDX-License-Identifier: GPL-3.0-or-later
// (c) 2021 Max Jonas Werner <mail@makk.es>

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	Notes string `json:"notes"`
}

func getItems(s string) ([]Object, error) {
	args := []string{"list", "items"}
	if s != "" {
		args = append(args, "--search", s)
	}
	cmd := exec.Command("bw", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	var obj []Object
	if err := json.Unmarshal(out.Bytes(), &obj); err != nil {
		return nil, err
	}
	return obj, nil
}

type DetailsDialog struct {
	*tview.Modal
	item           Object
	revealPassword bool
}

func NewDetailsDialog(pages *tview.Pages) *DetailsDialog {
	modal := tview.NewModal()
	dd := &DetailsDialog{
		Modal: modal,
	}
	modal.AddButtons([]string{"Close"}).
		SetDoneFunc(func(btnIdx int, btnLbl string) {
			dd.revealPassword = false
			pages.HidePage("dialog")
		}).
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyRune:
				switch event.Rune() {
				case 'r':
					dd.TogglePassword()
				}
			}
			return event
		})

	return dd
}

func (dd *DetailsDialog) SetItem(item Object) *DetailsDialog {
	dd.item = item
	dd.RenderCurrentItem()
	return dd
}

func (dd *DetailsDialog) RenderCurrentItem() {
	var buf strings.Builder
	buf.WriteString(dd.item.Name)
	if dd.item.Login != nil {
		buf.WriteString(fmt.Sprintf("\n\nUsername: %s\nPassword: ", dd.item.Login.Username))
		if dd.revealPassword {
			buf.WriteString(dd.item.Login.Password)
		} else {
			buf.WriteString("***")
		}
	}
	if dd.item.Notes != "" {
		buf.WriteString(fmt.Sprintf("\n\nNotes:\n%s", dd.item.Notes))
	}
	dd.SetText(buf.String())
}

func (dd *DetailsDialog) TogglePassword() {
	dd.revealPassword = !dd.revealPassword
	dd.RenderCurrentItem()
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
	pages := tview.NewPages()
	list := tview.NewList()

	detailsDialog := NewDetailsDialog(pages)

	bgColor := list.GetBackgroundColor()
	fgColor := list.GetBorderColor()
	list.SetBackgroundColor(fgColor)
	list.SetBorderColor(bgColor)
	list.SetMainTextColor(bgColor)
	list.SetSelectedBackgroundColor(bgColor)
	list.SetSelectedTextColor(fgColor)

	items, err := getItems(search)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed getting items from vault: %s\n", err.Error())
		os.Exit(1)
	}
	for _, item := range items {
		list.AddItem(item.Name, "", 0, nil)
	}
	list.ShowSecondaryText(false)
	list.SetHighlightFullLine(true)
	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		item := items[list.GetCurrentItem()]
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case 'p':
				if item.Login != nil {
					copyToClipboard(item.Login.Password)
				}
			case 'u':
				if item.Login != nil {
					copyToClipboard(item.Login.Username)
				}
			case 'q':
				app.Stop()
			}
		case tcell.KeyEnter:
			detailsDialog.SetItem(item)
			pages.ShowPage("dialog")
		}

		return event
	})

	pages.AddPage("list", list, true, true).
		AddPage("dialog", detailsDialog, true, false)

	if err := app.SetRoot(pages, true).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed running application: %s\n", err.Error())
		os.Exit(1)
	}
}
