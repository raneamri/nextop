package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

func writeInstanceConfig(instance Instance) error {
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
func healConfig() error {
	//fpath := "./gotop.conf"
	//parser, err := ioutil.ReadFile(fpath)

	var err error
	return err
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
