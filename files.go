package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

func writeConfig(instance Instance) error {
	/*
		Parse .conf to find connections section
	*/
	fpath := "gotop.conf"
	parser, err := ioutil.ReadFile(fpath)
	if err != nil {
		return err
	}
	connStart := strings.Index(string(parser), "[/connections]")
	connEnd := strings.Index(string(parser), "[/connections]")
	beforeSection := string(parser[:connStart])
	afterSection := string(parser[connEnd:])

	var (
		fdbms   string = "dbms=" + strdbms(instance.dbms) + " "
		fuser   string = "user=" + instance.user + " "
		fpass   string = "pass=" + string(instance.pass) + " "
		fport   string = "port=" + fmt.Sprint(instance.port) + " "
		fhost   string = "host=" + fmt.Sprint(instance.host) + " "
		fdbname string = "database-name=" + instance.dbname

		parsed []byte = []byte(beforeSection + fdbms + fuser + fpass + fport + fhost + fdbname + "\n" + afterSection)
	)

	err = ioutil.WriteFile(fpath, parsed, 0644)
	if err != nil {
		return err
	}

	return err
}

/*
Heals a file by iterating through its content and finding irregularities
*/
func healConfig() {
	//fpath := "./gotop.conf"
	//parser, err := ioutil.ReadFile(fpath)

}

/*
Deletes the config file and recreates it, then rewrites headers
*/
func resetConfig() {
	fpath := "gotop.conf"
	/*
		Recreate file
	*/
	os.Remove(fpath)
	os.Create(fpath)
	/*
		Re-write headers
	*/
	var headers = [2]string{"plugins", "connections"}
	for _, elem := range headers {
		var parsed []byte = []byte("[" + elem + "]\n\n[/" + elem + "]\n")
		ioutil.WriteFile(fpath, parsed, 0644)
	}
}

/*
Reads preset instances and puts them in slice by default
*/
func readConfig() ([]Instance, error) {
	var (
		instances []Instance
	)

	/*
		Same parsing as writeConfig
	*/
	fpath := "gotop.conf"
	parser, err := ioutil.ReadFile(fpath)
	if err != nil {
		return instances, err
	}

	/*
		Only keep 'connections' section
	*/
	connStart := strings.Index(string(parser), "[connections]")
	connEnd := strings.Index(string(parser), "[/connections]")
	config := string(parser[connStart+13 : connEnd-1])

	/*
		Parse each line
	*/
	lines := strings.Split(config, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || line == "\n" {
			continue
		}

		var inst Instance
		pairs := strings.Split(line, " ")
		for _, pair := range pairs {
			pair = strings.TrimSpace(pair)
			if pair == "" {
				continue
			}

			parts := strings.Split(pair, "=")
			if len(parts) != 2 {
				continue
			}

			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			switch key {
			case "dbms":
				inst.dbms = dbmsstr(value)
			case "user":
				inst.user = value
			case "pass":
				inst.pass = []byte(value)
			case "port":
				inst.port, _ = strconv.Atoi(value)
			case "host":
				inst.host = value
			case "database-name":
				inst.dbname = value
			}
		}

		instances = push_instance(instances, inst)
	}

	return instances, err
}

/*
Removes duplicates in config
*/
func cleanConfig() {
	fpath := "gotop.conf"
	parser, err := ioutil.ReadFile(fpath)
	if err != nil {
		panic(err)
	}
	config := string(parser)
	lines := strings.Split(config, "\n")
	uniquelines := make(map[string]struct{})
	ordered := []string{} // Slice to preserve order of unique lines

	for _, line := range lines {
		/*
			Skip empty lines
		*/
		if line == "" || line == "\n" {
			uniquelines[line] = struct{}{}
			ordered = append(ordered, line)
			continue
		}

		/*
			Check if line exists
		*/
		if _, exists := uniquelines[line]; exists {
			continue
		}

		/*
			Map line as unique
		*/
		uniquelines[line] = struct{}{}
		ordered = append(ordered, line) // Add line to the uniqueOrder slice
	}

	/*
		Format the unique lines
	*/
	output := strings.Join(ordered, "\n")

	/*
		Re-write config
	*/
	err = ioutil.WriteFile(fpath, []byte(output), 0644)
	if err != nil {
		panic(err)
	}
}

/*
Syncs []Instance slice to config
*/
func syncConfig(instances []Instance) {
	/*
		Write all instances in object to file
	*/
	for _, inst := range instances {
		writeConfig(inst)
	}
	/*
		Remove duplicates accross object & file
	*/
	cleanConfig()
	/*
		Put instances back in
	*/
	instances, _ = readConfig()
}
