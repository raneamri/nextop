package io

import (
	"os"

	"github.com/raneamri/nextop/errors"
	"github.com/raneamri/nextop/types"
	"github.com/raneamri/nextop/utility"
)

/*
	Handles startup arguments
*/

func ReadArgs(Instances map[string]types.Instance) {
	if len(os.Args) == 1 {
		return
	} else if len(os.Args) > 2 && len(os.Args) < 6 {
		var inst types.Instance

		inst.DBMS = utility.Dbmsstr(os.Args[1])
		if inst.DBMS == -1 {
			errors.ThrowArgError(os.Args)
		}
		inst.DSN = []byte(os.Args[2])

		if len(os.Args) > 3 {
			inst.ConnName = os.Args[3]
			if len(os.Args) > 4 {
				inst.Group = os.Args[4]
			} else {
				inst.Group = ""
			}
		} else if len(os.Args) < 4 {
			inst.ConnName = "unnamed"
		}

		if len(os.Args) == 5 {
			inst.Group = os.Args[4]
		}

		Instances[inst.ConnName] = inst
	} else {
		errors.ThrowArgError(os.Args)
	}
}
