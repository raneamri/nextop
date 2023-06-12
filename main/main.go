package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/term"
)

/*
View concurrencies
View plugins
*/

func main() {

	clearTerminal()

	/*
		Slice to store all instances
	*/
	instances, err := readConfig()
	if err != nil {
		catchConfigReadError(err, instances)
	}

	/*
		If user doesn't specify arguments on run
		prompt connection details and put in []Instance and/or .conf
		Note: os.Args[0] == main.go, so args @ index 1+

		If user specifies correct number of arguments, attempt parse
		If parsing fail, error is thrown

		If user specifies wrong number of arguments, exit with code 1
	*/
	if len(os.Args) == 1 {
		var loadnew string
		/*
			Prompt user if they want to use configged only if
			config isn't empty
		*/
		if len(instances) != 0 {
			fmt.Println("Load in with new instance?: [yes] ")
			fmt.Scanf("%s", &loadnew)
		} else {
			loadnew = "YES"
		}
		if fstr(loadnew) == "YES" {
			inst := newInstance()
			/*
				Write to config conditionally
			*/
			var write string
			fmt.Printf("Write to config?: [yes] ")
			fmt.Scanf("%s", &write)
			if fstr(write) == "YES" {
				/*
					Writes instance to .conf
				*/
				inst.db = connections.launchInstance(inst)
				instances = push_instance(instances, inst)
				err := writeConfig(inst)
				/*
					Program will attempt to heal config if error is thrown
					on fail, config file will be reset to default
					Healing preserves all unbroken configs while reset doesn't
					Note: heal unimplemented
				*/
				if err != nil {
					catchConfigWriteError(err, inst)
				}
			}

			/*
				Add instance to slice
			*/
			instances = push_instance(instances, inst)
		}
	} else if len(os.Args) > 3 && len(os.Args) < 8 {
		var inst Instance
		/*
			Unpack non-optional values
		*/
		inst.dbms = dbmsstr(os.Args[1])
		inst.user = os.Args[2]
		inst.pass = []byte(os.Args[3])

		/*
			Unpack optional values
		*/
		/*
			Flags to check which values have been found
		*/
		var (
			port   bool = false
			host   bool = false
			dbname bool = false
		)
		if len(os.Args) > 4 {
			for i := range os.Args {
				/*
					Check what arg. was given
					Contains . means host
					Fully numeric means port
					Else database name
				*/
				if regexp.MustCompile("^[0-9]+$").MatchString(os.Args[i]) {
					/*
						Given arg. = port
					*/
					if os.Args[i] != string(inst.pass) && os.Args[i] != string(inst.user) {
						num, err := strconv.Atoi(os.Args[i])
						if err != nil {
							throwArgError(os.Args)
						}
						inst.port = num
						port = true
					}
				} else if strings.Contains(os.Args[i], ".") {
					/*
						Given arg. = host
					*/
					inst.host = os.Args[i]
					host = true

				} else {
					/*
						Given arg. = dbname
						by deduction
					*/
					inst.dbname = os.Args[i]
					dbname = true
				}
			}
		}
		/*
			Catch any uninitialized arguments and set them
			Set localhost to 127.0.0.1
		*/
		if !port {
			inst.port = 3306
		}
		if !host {
			inst.host = "127.0.0.1"
		}
		if !dbname {
			inst.dbname = "none"
		}
		if inst.host == "localhost" {
			inst.host = "127.0.0.1"
		}
		inst.db = launchInstance(inst)

		instances = push_instance(instances, inst)
		syncConfig(instances)

	} else {
		throwArgError(os.Args)
	}

	/*
		Temporary system in the form of "game loop"
	*/
	var (
		fps      int
		interval time.Duration
	)
	fps = 60
	interval = time.Duration(fps/60) * time.Second
	for 1 == 1 {
		clearTerminal()
		dashboard := initDashboard(instances)
		fmt.Println(dashboard.String())

		time.Sleep(interval)
	}

}

/*
Prompts user to pick a DBMS
*/
func pickdbms() dbms_t {
	var input string
	fmt.Println("DBMS: ")
	fmt.Scanf("%s", &input)
	input = fstr(input)

	if input == "MYSQL" {
		return MYSQL
	} else if input == "ORACLE" {
		return ORACLE
	}

	fmt.Println("Unaccomodated/non-existant DBMS.")
	os.Exit(1)
	return -1
}

/*
Takes dbms_t and returns the dbms name as string
*/
func strdbms(dbms dbms_t) string {
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
func dbmsstr(dbms string) dbms_t {
	fstr(dbms)
	if dbms == "MYSQL" {
		return MYSQL
	} else if dbms == "ORACLE" {
		return ORACLE
	}

	return 0
}

/*
Creates a new instance and returns it
*/
func newInstance() Instance {
	var newInstance Instance
	newInstance.dbms = pickdbms()

	clearTerminal()
	fmt.Println("Enter username: ")
	fmt.Scanf("%s", &newInstance.user)

	clearTerminal()
	fmt.Print("Enter password: \n")
	password, _ := term.ReadPassword(int(syscall.Stdin))
	newInstance.pass = password

	clearTerminal()
	fmt.Println("Enter port (default:3306): ")
	fmt.Scanf("%d", &newInstance.port)
	if fmt.Sprint(newInstance.port) == "" {
		newInstance.port = 3306
	}

	clearTerminal()
	fmt.Println("Enter host (default:127.0.0.1): ")
	fmt.Scanf("%s", &newInstance.host)
	if newInstance.host == "" {
		newInstance.host = "127.0.0.1"
	}

	clearTerminal()
	fmt.Println("Enter database name (default:none): ")
	fmt.Scanf("%s", &newInstance.dbname)
	if newInstance.dbname == "" {
		newInstance.dbname = "none"
	}

	clearTerminal()
	return newInstance
}

/*
Pushes an item to the top of a slice
Also provides item with index
*/
func push_instance(instances []Instance, pushing Instance) []Instance {
	pushing.ndx = len(instances)
	instances = append(instances, pushing)
	return instances
}

/*
Removes an instance from []Instance by value
Re-indexes slice
*/
func pop_instance(instances []Instance, popping Instance) []Instance {
	var rm int = -1
	for i, it := range instances {
		if it.dbms == popping.dbms && it.host == popping.host && it.port == popping.port && it.user == popping.user {
			rm = i
		}
	}

	if rm != -1 {
		/*
			Create slice omitting rm element
		*/
		newinsts := make([]Instance, len(instances)-1)
		copy(newinsts[:rm], instances[:rm])
		copy(newinsts[rm:], instances[rm+1:])

		instances = newinsts
	} else {
		fmt.Println("Instance to pop unfound.")
	}

	/*
		Re-index instances
	*/
	for i, it := range instances {
		it.ndx = i
	}

	return instances
}
