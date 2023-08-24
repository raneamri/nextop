package io

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/raneamri/nextop/types"
	"github.com/raneamri/nextop/utility"
)

/*
Handles file i/o
*/

/*
Writes an instance to config
*/
func WriteConfig(instance types.Instance) {
	var (
		parser []byte

		connStart     int
		connEnd       int
		beforeSection string
		afterSection  string
	)

	const fpath string = ".nextop.conf"
	parser, _ = ioutil.ReadFile(fpath)

	connStart = strings.Index(string(parser), "[/connections]")
	connEnd = strings.Index(string(parser), "[/connections]")
	beforeSection = string(parser[:connStart])
	afterSection = string(parser[connEnd:])

	var (
		fdbms   string = "dbms=" + utility.Strdbms(instance.DBMS) + " "
		fdsn    string = "dsn=" + string(instance.DSN) + " "
		fdbname string = "conn-name=" + instance.ConnName + " "
		fgroup  string = "group=" + instance.Group

		parsed []byte = []byte(beforeSection + fdbms + fdsn + fdbname + fgroup + "\n" + afterSection)
	)

	ioutil.WriteFile(fpath, parsed, 0644)
}

/*
Reads preset instances and puts them in slice by default
*/
func ReadInstances(Instances map[string]types.Instance) {
	var (
		connStart int
		connEnd   int
		config    string
	)

	const fpath string = ".nextop.conf"
	contents, _ := ioutil.ReadFile(fpath)

	connStart = strings.Index(string(contents), "[connections]")
	connEnd = strings.Index(string(contents), "[/connections]")
	config = string(contents[connStart+13 : connEnd-1])

	var lines []string = strings.Split(config, "\n")
	for _, line := range lines {
		var (
			inst  types.Instance
			pairs []string
			parts []string
			key   string
			value string
		)

		line = strings.TrimSpace(line)
		if line == "" || line == "\n" {
			continue
		}

		pairs = strings.Split(line, " ")
		for _, pair := range pairs {
			pair = strings.TrimSpace(pair)
			if pair == "" {
				continue
			}

			parts = strings.Split(pair, "=")
			if len(parts) != 2 {
				continue
			}

			key = strings.TrimSpace(parts[0])
			value = strings.TrimSpace(parts[1])

			switch key {
			case "dbms":
				inst.DBMS = utility.Dbmsstr(value)
			case "dsn":
				inst.DSN = []byte(value)
			case "conn-name":
				inst.ConnName = value
			case "group":
				inst.Group = value
			}
		}

		Instances[inst.ConnName] = inst
	}
}

/*
Removes duplicates in config
*/
func CleanConfig() {
	var (
		config      string
		output      string
		lines       []string
		uniquelines map[string]struct{}
		ordered     []string
	)

	const fpath string = ".nextop.conf"
	parser, err := ioutil.ReadFile(fpath)
	if err != nil {
		panic(err)
	}
	config = string(parser)
	lines = strings.Split(config, "\n")
	uniquelines = make(map[string]struct{})

	for _, line := range lines {
		if line == "" || line == "\n" {
			uniquelines[line] = struct{}{}
			ordered = append(ordered, line)
			continue
		}

		if _, exists := uniquelines[line]; exists {
			continue
		}

		uniquelines[line] = struct{}{}
		ordered = append(ordered, line)
	}

	output = strings.Join(ordered, "\n")

	err = ioutil.WriteFile(fpath, []byte(output), 0644)
	if err != nil {
		panic(err)
	}
}

/*
Syncs []Instance slice to config
*/
func SyncConfig(Instances map[string]types.Instance) {
	for _, inst := range Instances {
		WriteConfig(inst)
	}

	CleanConfig()

	Instances = make(map[string]types.Instance)
	ReadInstances(Instances)
}

/*
Takes the value of the setting and returns its value as a string
*/
func FetchSetting(param string) string {
	var (
		settingsStart int
		settingsEnd   int
		config        string
		pairs         []string
		parts         []string
	)

	const fpath string = ".nextop.conf"
	contents, err := ioutil.ReadFile(fpath)
	if err != nil {
		panic(err)
	}

	settingsStart = strings.Index(string(contents), "[settings]")
	settingsEnd = strings.Index(string(contents), "[/settings]")
	config = string(contents[settingsStart+10 : settingsEnd-1])

	var lines []string = strings.Split(config, "\n")
	for _, line := range lines {
		var (
			value string
			key   string
		)

		line = strings.TrimSpace(line)
		if line == "" || line == "\n" {
			continue
		}

		pairs = strings.Split(line, " ")
		for _, pair := range pairs {
			pair = strings.TrimSpace(pair)
			if pair == "" {
				continue
			}

			parts = strings.Split(pair, "=")
			if len(parts) != 2 {
				continue
			}

			key = strings.TrimSpace(parts[0])
			value = strings.TrimSpace(parts[1])

			switch key {
			case param:
				return value
			}
		}
	}

	return string("-1")
}

/*
Writes processlist contents to a text file
*/
func ExportProcesslist(data *[]string) {
	var (
		fpath     string
		sb        strings.Builder
		re        *regexp.Regexp
		err       error
		out       string
		partition int
	)

	fpath = FetchSetting("export-path")
	re = regexp.MustCompile(`\s+`)
	curr := time.Now()

	out += "-- " + curr.Format("2006-01-02 15:04:05") + ":\n"

	file, err := os.OpenFile(fpath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	_, err = file.WriteString(out)
	if err != nil {
		panic(err)
	}

	for _, line := range *data {
		fmt.Println(len(line))
		line = re.ReplaceAllString(line, " ")

		for _, char := range line {
			sb.WriteRune(char)
			if char == ',' || char == ';' {
				sb.WriteRune('\n')
			}
		}

		out = sb.String()
		partition = strings.Index(out, "µ") + 3

		if out[len(out)-3] != 'µ' {
			out = out[:partition] + "\n" + out[partition:] + "\n\n"
		} else {
			out = out[:partition] + "\n" + "NO QUERY\n\n"
		}

		fmt.Println(len(out), "&", len(line))
		ch := fmt.Sprint(len(out)) + "&" + fmt.Sprint(len(line)) + "\n"
		_, err = file.WriteString(ch + out)
		if err != nil {
			panic(err)
		}

		sb.Reset()
	}
}
