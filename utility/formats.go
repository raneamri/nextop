package utility

import (
	"fmt"
	"regexp"
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
func FormatDuration(microseconds int64) string {

	var (
		out  float32 = 0
		unit string
	)

	switch {
	case microseconds <= 1e4:
		out = float32(microseconds) / 1e3
		unit = "ns"
	case microseconds <= 10e7:
		out = float32(microseconds) / 1e6
		unit = "μs"
	case microseconds <= 1e10:
		out = float32(microseconds) / 1e9
		unit = "ms"
	}

	ftime := fmt.Sprintf("%.4f"+unit, out)
	return ftime
}

func DeformatDuration(formattedDuration string) int64 {
	re := regexp.MustCompile(`^(\d+(\.\d+)?)\s*([nμm]s)$`)
	matches := re.FindStringSubmatch(formattedDuration)
	if len(matches) != 4 {
		return 0
	}

	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0
	}

	unit := matches[3]
	switch unit {
	case "ns":
		return int64(value * 1e3)
	case "μs":
		return int64(value * 1e6)
	case "ms":
		return int64(value * 1e9)
	}

	return 0
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

func ShuffleQuery(query types.Query, order []int) {
	if len(query.RawData) == 0 {
		return
	}
	maxIndex := len(query.RawData[0]) - 1
	for i, row := range query.RawData {
		newRow := make([]string, len(order))
		for j, idx := range order {
			if idx >= 0 && idx <= maxIndex && idx < len(row) {
				newRow[j] = row[idx]
			} else {
				newRow[j] = "n/a"
			}
		}
		query.RawData[i] = newRow
	}
}

func TrimNSprintf(format string, args ...interface{}) string {
	re := regexp.MustCompile(`\d+`)

	matches := re.FindAllString(format, -1)

	numbers := make([]int, len(matches))
	for i, match := range matches {
		num, err := strconv.Atoi(match)
		if err != nil {
			fmt.Printf("Error converting '%s' to int: %v\n", match, err)
			numbers[i] = 0
		} else {
			numbers[i] = num
		}
	}

	trimmedArgs := make([]interface{}, len(args))

	for i := 0; i < len(args); i++ {
		if i < len(numbers) && numbers[i] > 0 {
			if str, ok := args[i].(string); ok {
				if len(str) > numbers[i] {
					trimmedArgs[i] = str[:numbers[i]]
				} else {
					trimmedArgs[i] = str
				}
			} else {
				trimmedArgs[i] = args[i]
			}
		} else {
			trimmedArgs[i] = args[i]
		}
	}

	return fmt.Sprintf(format, trimmedArgs...)
}
