// SPDX-License-Identifier: GPL-3.0-or-later
// (c) 2021 Max Jonas Werner <mail@makk.es>

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Login struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Object struct {
	Type     string `json:"object"`
	Name     string `json:"name"`
	Login    *Login `json:"login"`
	Notes    string `json:"notes"`
	FolderId string `json:"folderId"`
	Folder   *Folder
}

func (o Object) String() string {
	if o.Folder != nil {
		return fmt.Sprintf("%s (%s)", o.Name, o.Folder.Name)
	}
	return o.Name
}

type Folder struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func getFolders() (map[string]*Folder, error) {
	cmd := exec.Command("bw", "list", "folders")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	var folders []*Folder
	if err := json.Unmarshal(out.Bytes(), &folders); err != nil {
		return nil, err
	}

	res := make(map[string]*Folder, len(folders))
	for _, f := range folders {
		res[f.ID] = f
	}

	return res, nil
}

func getItems(s string) ([]*Object, error) {
	folders, err := getFolders()
	if err != nil {
		return nil, err
	}

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
	var obj []*Object
	if err := json.Unmarshal(out.Bytes(), &obj); err != nil {
		return nil, err
	}
	for _, obj := range obj {
		if obj.FolderId != "" {
			obj.Folder = folders[obj.FolderId]
		}
	}
	return obj, nil
}

type DetailsDialog struct {
	*tview.Modal
	item           *Object
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

func (dd *DetailsDialog) SetItem(item *Object) *DetailsDialog {
	dd.item = item
	dd.RenderCurrentItem()
	return dd
}

func (dd *DetailsDialog) RenderCurrentItem() {
	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("%s\n\n", dd.item.Name))
	if dd.item.Folder != nil {
		buf.WriteString(fmt.Sprintf("Folder: %s\n", dd.item.Folder.Name))
	}
	if dd.item.Login != nil {
		buf.WriteString(fmt.Sprintf("Username: %s\nPassword: ", dd.item.Login.Username))
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

func handleFeedbackText(app *tview.Application, tv *tview.TextView) chan<- string {
	timer := time.NewTimer(0)
	in := make(chan string)
	go func() {
		for {
			select {
			case text := <-in:
				app.QueueUpdateDraw(func() {
					tv.SetText(text)
					timer.Reset(2 * time.Second)
				})
			case <-timer.C:
				app.QueueUpdateDraw(func() {
					tv.Clear()
				})
			}
		}
	}()

	return in
}

type KeyMappings struct {
	FocusFilter  rune
	ClearFilter  rune
	CopyPassword rune
	CopyUsername rune
	QuitApp      rune
	ListDown     rune
	ListUp       rune
}

var defaultKeyMappings = KeyMappings{
	FocusFilter:  '/',
	ClearFilter:  'c',
	CopyPassword: 'p',
	CopyUsername: 'u',
	QuitApp:      'q',
	ListDown:     'j',
	ListUp:       'k',
}

func main() {

	keyMappings := defaultKeyMappings

	items, err := getItems("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed getting items from vault: %s\n", err.Error())
		os.Exit(1)
	}

	app := tview.NewApplication()
	container := tview.NewFlex()
	container.SetDirection(tview.FlexRow)
	pages := tview.NewPages()

	feedbackDialog := tview.NewTextView().SetMaxLines(1).SetTextAlign(tview.AlignRight)
	feedbackCh := handleFeedbackText(app, feedbackDialog)

	detailsDialog := NewDetailsDialog(pages)
	list := tview.NewList()

	filteredItems := make([]*Object, len(items))
	copy(filteredItems, items)
	filterInput := tview.NewInputField()
	filterInput.
		SetChangedFunc(func(t string) {
			re, err := regexp.Compile("(?i)" + t)
			if err != nil {
				// Since search is performed while the user is typing, we need to expect the regexp compilation to fail.
				// That's why we just ignore any compilation error and simply exit early.
				return
			}
			filteredItems = make([]*Object, 0)
			for _, item := range items {
				if re.MatchString(item.Name) || (item.Folder != nil && re.MatchString(item.Folder.Name)) {
					filteredItems = append(filteredItems, item)
				}
			}
			list.Clear()
			for _, item := range filteredItems {
				list.AddItem(item.String(), "", 0, nil)
			}
		}).
		SetFinishedFunc(func(key tcell.Key) {
			app.SetFocus(list)
		})

	if len(os.Args) >= 2 {
		filterInput.SetText(os.Args[1])
	}

	container.AddItem(pages, 0, 1, true)
	container.AddItem(tview.NewFlex().
		AddItem(filterInput, 0, 1, false).
		AddItem(feedbackDialog, 0, 1, false), 1, 1, false)

	for _, item := range filteredItems {
		list.AddItem(item.String(), "", 0, nil)
	}
	list.ShowSecondaryText(false)
	list.SetHighlightFullLine(true)
	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		var item *Object
		if len(filteredItems) > 0 {
			item = filteredItems[list.GetCurrentItem()]
		}
		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case keyMappings.ClearFilter:
				filterInput.SetText("")
			case keyMappings.ListUp:
				list.SetCurrentItem((list.GetItemCount() + list.GetCurrentItem() - 1) % list.GetItemCount())
			case keyMappings.ListDown:
				list.SetCurrentItem((list.GetCurrentItem() + 1) % list.GetItemCount())
			case keyMappings.FocusFilter:
				app.SetFocus(filterInput)
			case keyMappings.CopyPassword:
				if item != nil && item.Login != nil {
					copyToClipboard(item.Login.Password)
					feedbackCh <- "\npassword copied to clipboard"
				}
			case keyMappings.CopyUsername:
				if item != nil && item.Login != nil {
					copyToClipboard(item.Login.Username)
					feedbackCh <- "\nusername copied to clipboard"
				}
			case keyMappings.QuitApp:
				app.Stop()
			}
		case tcell.KeyEnter:
			if item != nil {
				detailsDialog.SetItem(item)
				pages.ShowPage("dialog")
			}
		}

		return event
	})

	pages.AddPage("list", list, true, true).
		AddPage("dialog", detailsDialog, true, false)

	if err := app.SetRoot(container, true).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "failed running application: %s\n", err.Error())
		os.Exit(1)
	}
}
