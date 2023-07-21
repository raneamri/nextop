# Nextop
Nextop is a powerful program designed for monitoring MySQL and other Database Management Systems (DBMSs), developed during a summer internship in 2023. This versatile tool provides support for plugin creation, allowing users to extend its functionality according to their specific needs.

## Startup
To build Nextop, execute the following command:
```bash
    make build
```

Once built, you can run Nextop with the following command:
```bash
    ./nextop
```

Alternatively, you can provide arguments to customize the behavior:
```bash
    ./nextop <dbms> <dsn> <conn-name> <group>
```
Please note that on Mac M1 zsh, it is necessary to use quotation marks around the DSN. The last argument is optional. Although naming your connections is not mandatory, we strongly recommend giving them unique names, especially when monitoring multiple connections. Grouping connections is also an optional feature that allows you to manage them as a unit.

Upon establishing a successful connection, Nextop will create an instance in the configuration for easy future access.

## Interface
### Controls
View menu.

### Configurations
When running Nextop for the first time, without any arguments, or without any pre-configured connections, you will be directed to the configuration page:

![ConfigPage](https://github.com/raneamri/nextop/blob/main/img/config.png)

We highly recommend naming connections and giving them unique identifiers. Grouping connections facilitates managing them efficiently. Currently, to remove connections or modify configurations, you need to access the nextop.conf file manually. After submitting changes, the program will attempt to connect with the specified DSN.

*As of Fri 23rd June 2023, Nextop supports MySQL as the only DBMS.

### Processlist
The processlist dynamically displays ongoing processes, with different colors indicating their latency:
- 5 seconds -> blue
- 10 seconds -> yellow
- 30 seconds -> red
- 1 minute -> dark red

![Processlist](https://github.com/raneamri/nextop/blob/main/img/processlist.png)

The "Filters" section in the top left corner enables users to omit unwanted messages or search for specific messages based on substrings. The case sensitivity of filters can be adjusted in configurations. To filter for multiple items, separate them with commas.

A group filter is available below the filters, allowing users to display processes specific to a particular group.

To clear all filters quickly, press backslash "\", and to remove only group filters, use the forward slash "/".

### InnoDB Dashboard
The InnoDB Dashboard provides essential data from the InnoDB engine, including pie charts that represent statistics. If a pie chart does not render and shows 0%, it may be an issue with termdash or the program, and we are actively working to resolve it.

![InnoDBDashboard](https://github.com/raneamri/nextop/blob/main/img/innodb.png)

### Memory Dashboard
The Memory Dashboard presents memory-related statistics.

![MemoryDashboard](https://github.com/raneamri/nextop/blob/main/img/memory.png)

### Error Log

![ErrorLog](https://github.com/raneamri/nextop/blob/main/img/error.png)

### Lock Log


### Plugins
Coming less soon.