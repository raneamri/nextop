package utility

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/raneamri/nextop/types"
)

/*
Formats a string into block capitals and turns spaces into underscores
*/
func Fstr(formattable string) string {
	words := strings.Fields(formattable)
	formattedWords := make([]string, len(words))

	for i, word := range words {
		upperCaseWord := strings.Map(unicode.ToUpper, word)
		formattedWords[i] = strings.ReplaceAll(upperCaseWord, " ", "_")
	}

	return strings.Join(formattedWords, "_")
}

/*
Formats time as float64 seconds into hours mins secs
*/
func Ftime(duration int) string {
	/*
		float64 -> time.Duration
	*/
	fduration := time.Duration(duration * int(time.Second))

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
Formats time as int64 picoseconds to string ms
*/
func FpicoToMs(duration int64) string {
	/*
		int64 -> time.Duration
	*/
	var ms float32 = float32(duration) / 1e9

	/*
		Format and concatenate
	*/
	ftime := fmt.Sprintf("%.4fms", ms)
	return ftime
}

/*
Same as FpicoToMs but to microseconds (µs)
*/
func FpicoToUs(duration int64) string {
	/*
		int64 -> time.Duration
	*/
	us := duration / int64(time.Microsecond)

	/*
		Format and concatenate
	*/
	ftime := fmt.Sprintf("%dµs", us)
	return ftime
}

/*
Takes dbms_t and returns the dbms name as string
^ADD YOUR DBMS HERE
*/
func Strdbms(dbms types.DBMS_t) string {
	switch dbms {
	case types.MYSQL:
		return "mysql"
	case types.POSTGRES:
		return "postgres"
	default:
		return "n/a"
	}
}

/*
Inverse function to strdbms
Takes string and converts to dbms_t
Returns -1 if dbms is invalid
*/
func Dbmsstr(dbms string) types.DBMS_t {
	switch Fstr(dbms) {
	case "MYSQL":
		return types.MYSQL
	case "POSTGRES":
		return types.POSTGRES
	default:
		return -1
	}
}

/*
Converts bytes as one integer to MiB with prefix
*/
func BytesToMiB(bytes int) string {
	mb := bytes / (1024 * 1024)
	fmb := fmt.Sprint(mb) + " MiB"

	return fmb
}

func Fnum(num int) string {
	numStr := strconv.Itoa(num)
	length := len(numStr)

	if length <= 3 {
		return numStr
	}

	var formattedNum strings.Builder

	firstPartLength := length % 3
	if firstPartLength == 0 {
		firstPartLength = 3
	}

	formattedNum.WriteString(numStr[:firstPartLength])

	for i := firstPartLength; i < length; i += 3 {
		formattedNum.WriteByte(',')
		formattedNum.WriteString(numStr[i : i+3])
	}

	return formattedNum.String()
}

/*
Takes in string name and returns state type
*/
func Statestr(str string) types.State_t {
	switch str {
	case "MENU":
		return types.MENU
	case "PROCESSLIST":
		return types.PROCESSLIST
	case "THREAD_ANALYSIS":
		return types.THREAD_ANALYSIS
	case "DB_DASHBOARD":
		return types.DB_DASHBOARD
	case "MEM_DASHBOARD":
		return types.MEM_DASHBOARD
	case "ERR_LOG":
		return types.ERR_LOG
	case "LOCK_LOG":
		return types.LOCK_LOG
	case "REPLICATION":
		return types.REPLICATION
	case "TRANSACTIONS":
		return types.TRANSACTIONS
	case "CONFIGS":
		return types.CONFIGS
	case "QUIT":
		return types.QUIT
	default:
		return types.MENU
	}
}
