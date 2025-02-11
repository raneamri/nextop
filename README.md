# Nextop
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/raneamri/nextop)
![Static Badge](https://img.shields.io/badge/interface-termdash-red)
![Static Badge](https://img.shields.io/badge/version-0.0.3-yellow)
![Static Badge](https://img.shields.io/badge/development-active-blue)
![GitHub repo size](https://img.shields.io/github/repo-size/raneamri/nextop)
![GitHub Repo stars](https://img.shields.io/github/stars/raneamri/nextop)

Nextop is a lightweight program designed for monitoring MySQL, (and in the future) PostgreSQL and other Database Management Systems (DBMS). This versatile tool aims to provide support for plugin creation, allowing users to extend its functionality according to their specific needs.

---------------------------------------
* [Startup](#startup)
    * [Parameters](#parameters)
    * [Interface](#interface)
        * [Controls](#controls)
        * [Configurations](#configurations)
        * [Processlist](#processlist)
        * [Thread Analysis](#thread-analysis)
        * [InnoDB Dashboard](#innodb-dashboard)
        * [Memory Dashboard](#memory-dashboard)
        * [Error Log](#error-log)
    * [Backend](#back-end)
    * [License](#license)

---------------------------------------

## Startup
To build `Nextop`, execute the following command:
```bash
    make build
```

Once built, you can run Nextop with the following command:
```bash
    ./nextop
```

Alternatively, you can provide arguments to customize the behavior:
```bash
    ./nextop <dbms> "<dsn>" <conn-name> <group>
```

Upon establishing a successful connection, Nextop will create an instance in the configuration for easy future access.

## Parameters
Some parameters are more important than others. These are the ones I believe weigh most.

`startup-view`
```
Type:           type.State_t

Valid Values: 	MENU
                PROCESSLIST
                THREAD_ANALYSIS
                DB_DASHBOARD
                MEM_DASHBOARD
                ERR_LOG
                LOCK_LOG
                REPLICATION
                TRANSACTIONS
                CONFIGS

Default:        PROCESSLIST
```
startup-view chooses what you see the moment you launch nextop (as long as a valid connection is provided).

`refresh-rate`
```
Type:         time.Milliseconds

Valid Values: [10, ∞)

Default:      1,000
```
Tested with a locally hosted database, `Nextop` manages >10ms well.

If it seems some data is failing to display, adjust the refresh-rate up. As mentioned in `Backend`, `Nextop` is a best-effort system and doesn't guarantee for all queries to be executed in full if rushed.


## Interface
### Controls
`Interfaces`
--------------------------------------------------------------------
* `?`: Help / Menu
* `P`: Processlist
* `D`: InnoDB Dashboard
* `M`: Memory Dashboard
* `E`: Error Log
* `L`: Lock Log
* `R`: Replication
* `T`: Transactions
* `C`: Configs
* `Q`: Quit
--------------------------------------------------------------------

`Misc.`
--------------------------------------------------------------------
* `ESC`: Previous view
* `TAB`: Reload page
* `->`: Cycle to next connection (first if no next)
* `<-`: Cycle to previous connection (last if no previous)
* `\`: Quick clear filters
* `/`: Clear group filters
* `=`: Pause
* `+`: Increase refresh rate by 100ms
* `-`: Decrease refresh rate by 100ms
* `_`: Export processlist (works only if view is paused)
--------------------------------------------------------------------

### Configurations
When running Nextop for the first time, without any arguments, without any pre-configured connections, or with no active configured connection, you will be directed to the configuration page:

![ConfigPage](https://github.com/raneamri/nextop/blob/main/img/config.png)

Naming connections and giving them unique identifiers facilitates managing them efficiently. Currently, to remove connections or modify configurations, you need to access the .nextop.conf file manually. After submitting changes, the program will attempt to connect with the specified DSN.

### Processlist
The processlist dynamically displays ongoing processes, with different colors indicating their latency:

--------------------------
- `5ms   -> yellow`
- `10ms  -> red`
- `50ms  -> dark red`
- `100ms -> magenta`
--------------------------

![Processlist](https://github.com/raneamri/nextop/blob/main/img/processlist.png)

The `Processlist`, unlike other interfaces, queries all active connections for their processlists and aggregates them, regardless of DBMS.

The `Filters` allows omitting of unwanted messages or search for specific messages based on substrings. The case sensitivity of filters can be adjusted in configurations. To filter for multiple items, separate them with commas.

The `Kill` textbox allows killing a connection by ID. Use at your own risk, killing important connections can interfere with the program.

The `Analyse` textbox accepts a thread ID, which it'll then explain the query in that thread if possible.

The `Processlist` can be paused & exported to a self contained .sql file.

### Thread Analysis
`Thread Analysis` will `EXPLAIN` any query running in a given thread and display the details.

### InnoDB Dashboard
The `InnoDB Dashboard` provides essential data from the InnoDB engine.

![InnoDBDashboard](https://github.com/raneamri/nextop/blob/main/img/innodb.png)

### Memory Dashboard
The `Memory Dashboard` presents:
* Code-base specific allocation
* User allocation
* Global allocation
* Hardware allocation

And a linechart to illustrate global allocation.

![MemoryDashboard](https://github.com/raneamri/nextop/blob/main/img/memory.png)

### Error Log
The `Error Log` page comes equipped with a filter similar to the processlist.

Errors are categorised four-ways:
* Warning (yellow)
* System (blue)
* Error (red)
* Other

The number of each type of error on-screen is presented on a linechart.

![ErrorLog](https://github.com/raneamri/nextop/blob/main/img/error.png)

## Back-End
Nextop uses a best-effort periodic collect and display at interval system, meaning the program will attempt to fetch data from all connections, and all data that has been fetched within a customisable "refresh-rate" period will be displayed at the end of it. This means the end user may have to adjust the refresh rate if some connections aren't pinged in time.

To facilitate fetching long queries without the use of file reading or taxing memory, the program fetches queries using "keywords" which are keys to maps that store pointers to functions that return a query. An example of these maps in use is:

```bash
lookup map[string]func() string = GlobalQueryMap[Instances[conn].DBMS]
query, _ = queries.Query(..., lookup["processlist"]())

/*
    Alternatively
*/

query, _ = queries.Query(..., GlobalQueryMap["mysql"]["processlist"]())
```

This system also facilitates the implementation of new DBMSs, by simply allowing the end user to add a map for it and a relevant .go file to store its queries as string returning functions.

To overcome the hurdle of non-homogenous cardinality in otherwise equivalent schemas among DBMSs, nextop uses ordinality to describe synonymy. Every return from a query is bound to a slice of unique indexes, which bind values to a field, enabling union of these schemas. Nextop was designed in the image of MySQL, and so fields that may not exist for other DBMSs are managed by this system also, where any blank field is replaced with "n/a" if unfetchable.

The interface is managed by a "simple C++ game"-type state machine and the display is managed by the github.com/termdash library.