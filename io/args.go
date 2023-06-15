package io

import (
	"os"

	"github.com/raneamri/gotop/errors"
	"github.com/raneamri/gotop/types"
	"github.com/raneamri/gotop/utility"
)

func ReadArgs(instances []types.Instance) []types.Instance {
	if len(os.Args) == 3 || len(os.Args) == 4 {
		var inst types.Instance
		/*
			Unpack non-optional values
		*/
		inst.DBMS = utility.Dbmsstr(os.Args[1])
		if inst.DBMS == -1 {
			errors.ThrowArgError(os.Args)
		}
		inst.DSN = os.Args[2]

		if len(os.Args) == 4 {
			inst.Dbname = os.Args[3]
		} else {
			inst.Dbname = "<unnamed>"
		}

		instances = utility.PushInstance(instances, inst)
	} else {
		errors.ThrowArgError(os.Args)
	}

	return instances
}
