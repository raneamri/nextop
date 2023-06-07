package main

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
	"unicode"

	_ "github.com/go-sql-driver/mysql"
)

/*
Clears stdin
*/
func clearTerminal() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}

	cmd.Stdout = os.Stdout
	cmd.Run()
}

/*
Formats a string into block capitals and turns spaces into underscores
*/
func fstr(formattable string) string {
	words := strings.Fields(formattable) // Split the string into words
	formattedWords := make([]string, len(words))

	for i, word := range words {
		upperCaseWord := strings.Map(unicode.ToUpper, word)             // Convert word to block capitals
		formattedWords[i] = strings.ReplaceAll(upperCaseWord, " ", "_") // Replace spaces with underscores
	}

	return strings.Join(formattedWords, "_")
}
