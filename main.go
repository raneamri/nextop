package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/raneamri/gotop/errors"
	"github.com/raneamri/gotop/io"
	"github.com/raneamri/gotop/services"
	"github.com/raneamri/gotop/types"
	"github.com/raneamri/gotop/ui"
	"github.com/raneamri/gotop/utility"

	_ "github.com/go-sql-driver/mysql"
)

/*
View concurrencies
View plugins
*/

func main() {

	utility.ClearTerminal()

	/*
		Slice to store all instances
	*/
	instances, err := io.ReadConfig()
	if err != nil {
		errors.CatchConfigReadError(err, instances)
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
		if utility.Fstr(loadnew) == "YES" {
			inst := io.NewInstance()
			/*
				Write to config conditionally
			*/
			var write string
			fmt.Printf("Write to config?: [yes] ")
			fmt.Scanf("%s", &write)
			if utility.Fstr(write) == "YES" {
				/*
					Writes instance to .conf
				*/
				inst.DB = services.LaunchInstance(inst)
				instances = utility.PushInstance(instances, inst)
				err := io.WriteConfig(inst)
				/*
					Program will attempt to heal config if error is thrown
					on fail, config file will be reset to default
					Healing preserves all unbroken configs while reset doesn't
					Note: heal unimplemented
				*/
				if err != nil {
					errors.CatchConfigWriteError(err, inst)
				}
			}

			/*
				Add instance to slice
			*/
			instances = utility.PushInstance(instances, inst)
		}
	} else if len(os.Args) > 3 && len(os.Args) < 8 {
		var inst types.Instance
		/*
			Unpack non-optional values
		*/
		inst.DBMS = utility.Dbmsstr(os.Args[1])
		inst.User = os.Args[2]
		inst.Pass = []byte(os.Args[3])

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
					if os.Args[i] != string(inst.Pass) && os.Args[i] != string(inst.User) {
						num, err := strconv.Atoi(os.Args[i])
						if err != nil {
							errors.ThrowArgError(os.Args)
						}
						inst.Port = num
						port = true
					}
				} else if strings.Contains(os.Args[i], ".") {
					/*
						Given arg. = host
					*/
					inst.Host = os.Args[i]
					host = true

				} else {
					/*
						Given arg. = dbname
						by deduction
					*/
					inst.Dbname = os.Args[i]
					dbname = true
				}
			}
		}
		/*
			Catch any uninitialized arguments and set them
			Set localhost to 127.0.0.1
		*/
		if !port {
			inst.Port = 3306
		}
		if !host {
			inst.Host = "127.0.0.1"
		}
		if !dbname {
			inst.Dbname = "none"
		}
		if inst.Host == "localhost" {
			inst.Host = "127.0.0.1"
		}
		inst.DB = services.LaunchInstance(inst)

		instances = utility.PushInstance(instances, inst)
		io.SyncConfig(instances)

	} else {
		errors.ThrowArgError(os.Args)
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
		utility.ClearTerminal()
		dashboard := ui.InitDashboard(instances)
		fmt.Println(dashboard.String())

		time.Sleep(interval)
	}

}
