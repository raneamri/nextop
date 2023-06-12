package main

import (
	"fmt"
	"strings"
	"time"
	"unicode"
)

/*
Formats a string into block capitals and turns spaces into underscores
*/
func Fstr(formattable string) string {
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
func Ftime(duration float64) string {
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

/*
Takes dbms_t and returns the dbms name as string
*/
func Strdbms(dbms DBMS_t) string {
	if dbms == MYSQL {
		return "MYSQL"
	} else if dbms == ORACLE {
		return "ORACLE"
	}

	/*
		Should never be reached considering previous checks
	*/
	return ""
}

/*
Inverse function to strdbms
Takes string and converts to dbms_t
*/
func Dbmsstr(dbms string) DBMS_t {
	Fstr(dbms)
	if dbms == "MYSQL" {
		return MYSQL
	} else if dbms == "ORACLE" {
		return ORACLE
	}

	return 0
}
