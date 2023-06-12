package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
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

/*
Formats time as float64 seconds into hours mins secs
*/
func ftime(duration float64) string {
	/*
		float64 -> time.Duration
	*/
	fduration := time.Duration(duration * float64(time.Second))

	/*
		Extract hours, minutes & seconds
	*/
	hours := int(fduration.Hours())
	minutes := int(fduration.Minutes()) % 60
	seconds := int(fduration.Seconds()) % 60

	/*
		Format and concatenate
	*/
	ftime := fmt.Sprintf("%dh %dmin %ds", hours, minutes, seconds)
	return ftime
}
