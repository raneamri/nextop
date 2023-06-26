# GoTop (Pre-Release)
Innotop for MySQL (& other DBMSs) written in GoLang.
Developed over the course of my 2023 summer internship.

Project name still undecided.

## Startup
GoTop can be built using the command:
```bash
    make build
```

and can be called using the command:
```bash
    ./gotop
```

on Mac or 
```bash
    gotop
```

on Windows

or called with arguments:
```bash
    ./gotop <dbms> <dsn> <conn-name>
```
where quotation marks around the DSN is necessary on Mac M1 zsh and the last argument is optional.
While naming your connection isn't required, it is greatly recommended to name connections if you're monitoring multiple.

Upon successful connection, instance will be written in config for easy access in the future.

## Interface
### Controls
Will add

### Configs
If gotop is called for the first time with no arguments, you will be sent directly to this page:

![ConfigPage](https://github.com/raneamri/gotop/blob/main/img/config.png)

Currently, to remove connections or change configurations, you will have to access gotop.conf and do it manually.
Upon submission, the program will attempt to connect with the specified DSN.

*As of Fri 23rd June 2023, the only DBMS supported is MySQL.

### Processlist
The processlist will dynamically show you ongoing processes. Processes will appear in different colors depending on their latency:
- 5s = blue
- 10s = yellow
- 30s = red
- 1min = dark red

![Processlist](https://github.com/raneamri/gotop/blob/main/img/processlist.png)

Filtering for processlist coming in July.

### InnoDB Dashboard
InnoDB Dashboard primarily shows data from the InnoDB engine.
Note that if a pie chart doesn't render, it is trying to show a value of 0%.
I'm not sure yet if this is an issue with termdash or my program but I'm looking to fix this soon.

![InnoDBDashboard](https://github.com/raneamri/gotop/blob/main/img/innodb.png)

### Memory Dashboard
![MemoryDashboard](https://github.com/raneamri/gotop/blob/main/img/memory.png)

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

![ErrorLog](https://github.com/raneamri/gotop/blob/main/img/error.png)

### Lock Log
Coming very soon!

### Plugins
Coming soon enough.