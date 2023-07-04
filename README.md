# NEXTOP (Pre-Release)
Innotop for MySQL (& other DBMSs) written in GoLang.
Developed over the course of my 2023 summer internship.

Project name still undecided.

## Startup
NEXTOP can be built using the command:
```bash
    make build
```

and can be called using the commands:
```bash
    ./nextop
```

or called with arguments:
```bash
    ./nextop <dbms> <dsn> <conn-name> <group>
```
where quotation marks around the DSN is necessary on Mac M1 zsh and the last argument is optional.
While naming your connection isn't required, it is greatly recommended to name connections if you're monitoring multiple.
Grouping is also optional, and not grouping your connections has no repercussions.

Upon successful connection, instance will be written in config for easy access in the future.

## Interface
### Controls
Will add

### Configs
If NEXTOP is called for the first time with no arguments, you will be sent directly to this page:

![ConfigPage](https://github.com/raneamri/nextop/blob/main/img/config.png)

It is strongly advised to name connections and to give them unique names.
Grouping allows relation of multiple connections as to be able manage them as a unit.

Currently, to remove connections or edit configurations, you will have to access nextop.conf and do it manually.
Upon submission, the program will attempt to connect with the specified DSN.

*As of Fri 23rd June 2023, the only DBMS supported is MySQL.

### Processlist
The processlist will dynamically show you ongoing processes. Processes will appear in different colors depending on their latency:
- 5 seconds -> blue
- 10 seconds -> yellow
- 30 seconds -> red
- 1 minute -> dark red

![Processlist](https://github.com/raneamri/nextop/blob/main/img/processlist.png)

The top left section "Filters" allows the user to ommit unwanted messages and/or find specific messages.
The way the filter works is by looking for the entered substring in the "Message".

*Group management coming soon

### InnoDB Dashboard
InnoDB Dashboard primarily shows data from the InnoDB engine.
Note that if a pie chart doesn't render, it is trying to show a value of 0%.
I'm not sure yet if this is an issue with termdash or my program but I'm looking to fix this soon.

![InnoDBDashboard](https://github.com/raneamri/nextop/blob/main/img/innodb.png)

### Memory Dashboard
![MemoryDashboard](https://github.com/raneamri/nextop/blob/main/img/memory.png)

### Error Log
In the top right section of the interface "Statistics" is a line graph which will show change in log types, where:
- blue line = system count
- yellow line = warning count
- red line = error count
This data is non-historical and the graph is reset upon leaving the page.

The bottom section "Log" displays all retrieved logs from the database and color codes them.

The top left section "Filters" allows the user to ommit unwanted messages and/or find specific messages.
The way the filter works is by looking for the entered substring in the "Message".

As of now, the filters ignore timestamp and thread. This is on purpose but subject to change based on feedback.

![ErrorLog](https://github.com/raneamri/nextop/blob/main/img/error.png)

### Lock Log
Coming very soon!

### Plugins
Coming soon enough.