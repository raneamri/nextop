## PLUGINS
NOTE THAT THE PLUGIN SYSTEM IS YET TO BE TESTED AND THAT NEXTOP DOES NOT
CURRENTLY FEATURE SUPPORT FOR PLUGINS. THIS WILL HAVE CHANGED WHEN THE SOFTWARE
GETS ITS PROPER LICENSE, CODE OF CONDUCT AND RELEASE.

Nextop is a modular & open-source software, making it easy to build plugins for.
Go is a statically typed language, and so while it isn't impossible to make a fully automatic
plugin system, it would probably take longer than it took to develop the project and 
wouldn't be worthwhile. Besides, it's already quite easy. Make sure to read this markdown thoroughly.

Tip: to easily find where the guide asks you to locate, using CTRL+F / CMD+F and search for "^".

# LIMITATIONS
Nextop is limited to SQL languages.

Some parts of nextop's default variables don't exist in other languages since it was originally built for MySQL.
This won't break plugins, but those variables will show up as n/a or blank unless accomodated for.
Do note that most variables will translate over.

While this program takes constructive feedback from users, plugins cannot fundamentally change the
program, only add to it. To improve on the program, please make a pull request.

# WALKTHROUGH
Before starting, make sure to consult the CODE OF CONDUCT.

First, establish what type of plugin you're making. Is it adding support for a DBMS? Is it adding
interface pages?


# DBMS
Find DBMS_t in types/types.go and insert the dbms you're adding at the bottom:
```bash
const (
	MYSQL DBMS_t = iota
	POSTGRE
    MARIADB
)
```

Then, locate Strdbms() & Dbmsstr() in utility/formats.go and add your dbms to the selections:
```bash
    func Strdbms(dbms types.DBMS_t) string {
        switch dbms {
        case types.MYSQL:
            return "mysql"
        case types.POSTGRE:
            return "postgre"
        case types.MARIADB:
            return "mariadb"
        default:
            return "n/a"
        }
    }

    func Dbmsstr(dbms string) types.DBMS_t {
        dbms = Fstr(dbms)
        switch dbms {
        case "MYSQL":
            return types.MYSQL
        case "POSTGRE":
            return types.POSTGRE
        case "MARIADB:
            return types.MARIADB
        default:
            return -1
        }
    }
```

Next, find the global variables in ui/state_machine.go and add a map to store your queries:
```bash
...
	GlobalQueryMap map[types.DBMS_t]map[string]func() string = make(map[types.DBMS_t]map[string]func() string)
	MySQLQueries   map[string]func() string                  = make(map[string]func() string)
    MariaDBQueries map[string]func() string                  = make(map[string]func() string)
```

The second last step is to locate the directory queries/ and adding a file for your plugin.
Mine would be mariadb_queries.go.

In dict.go, you will find a list of comprehensive keys that the program uses to locate queries.
Go in your newly made file and create a function dictionary:
```bash
    func MariaDBFuncDict() []func() string {
        return []func() string{MariaDBProcesslistQuery,
                               MariaDBUptimeQuery,
                               ...
                               MariaDBLocksQuery,
        }
    }
```

Make sure that for each keyword in the keys list, there is a matching pointer to a function that 
returns a query at the same index.

Lastly, write all your queries in functions matching those named in your function dictionary.
If this confuses you, visit mysql_queries.go.

## CONCLUSION
After this your plugin should be complete and ready for deployment. Please submit a pull request
once you are sure it works hand in hand with the program for your plugin to be added to the cannonical
version of the software.