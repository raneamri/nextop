package io

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/raneamri/gotop/errors"
	"github.com/raneamri/gotop/services"
	"github.com/raneamri/gotop/types"
	"github.com/raneamri/gotop/utility"
)

func ReadArgs(instances []types.Instance) {
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
			inst := NewInstance()
			/*
				Write to config conditionally
			*/
			var write string
			fmt.Printf("Write to config?: [yes] ")
			fmt.Scanf("%s", &write)
			if utility.Fstr(write) == "YES" {
				/*
					Start driver
				*/
				services.LaunchInstance(inst)
				/*
					Store config in dynamic slice
				*/
				instances = utility.PushInstance(instances, inst)
				/*
					Syncs dynamically stored configs to statically stored configs
					Syncing involves writing to config (view files.go)
				*/
				SyncConfig(instances)
			}
		} else {
			return
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
				} else if strings.Contains(os.Args[i], ".") && os.Args[i] != "./gotop" {
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

		inst.Driver = nil

		instances = utility.PushInstance(instances, inst)
		SyncConfig(instances)

	} else {
		errors.ThrowArgError(os.Args)
	}
}
