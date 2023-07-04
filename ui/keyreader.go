package ui

import (
	"context"

	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/raneamri/nextop/errors"
	"github.com/raneamri/nextop/io"
	"github.com/raneamri/nextop/types"
)

/*
	Should've been in io package but placed here to avoid
	circular dependency and for ease of access to global variables
*/

func KeyList() []string {
	return []string{"menu", "processlist", "dbdashboard", "memdashboard", "errlog",
		"locklog", "configs", "back", "quit"}
}

func ValidateKeybinds() {
	var (
		keylist  []string = KeyList()
		keybinds []string
		dupes    map[string]bool = make(map[string]bool)
	)

	for _, key := range keylist {
		keybinds = append(keybinds, io.FetchKeybind(key))
	}

	for _, key := range keybinds {
		if !dupes[key] {
			dupes[key] = true
		} else {
			errors.ThrowKeybindError(key)
		}
	}
}

func FcustomKey() {

}

func FetchKeyreader(cancel context.CancelFunc) func(k *terminalapi.Keyboard) {
	var (
		keylist  []string = KeyList()
		keybinds []string
	)

	for _, key := range keylist {
		keybinds = append(keybinds, io.FetchKeybind(key))
	}

	var keyreader func(k *terminalapi.Keyboard) = func(k *terminalapi.Keyboard) {
		switch k.Key {
		case 'p', 'P':
			State = types.PROCESSLIST
			cancel()
		case 'd', 'D':
			State = types.DB_DASHBOARD
			cancel()
		case 'm', 'M':
			State = types.MEM_DASHBOARD
			cancel()
		case 'e', 'E':
			State = types.ERR_LOG
			cancel()
		case 'l', 'L':
			State = types.LOCK_LOG
			cancel()
		case 'c', 'C':
			State = types.CONFIGS
			cancel()
		case keyboard.KeyEsc:
			State = Laststate
			cancel()
		case 'q', 'Q':
			State = types.QUIT
			cancel()
		}
	}

	return keyreader
}
