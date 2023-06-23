# GoTop (EARLY-DEV)
Innotop for MySQL (& other DBMSs) written in GoLang.
Developed over the course of my 2023 summer internship.

## Startup
GoTop can be built using the command:
```bash
    make build
```

while in the correct directory & can be called using the command:
```bash
    ./gotop
```

or called with arguments:
```bash
    ./gotop <dbms> "<dsn>" <conn-name>
```
where the quotation marks around the DSN is necessary on Mac M1 zsh and the last argument is completely optional.
While naming your connection isn't required, it is greatly recommended to name connections if you're monitoring multiple.

Upon successful connection, instance will be written in config for easy access in the future.

## Interface
### Controls
Will add

### Configs
If gotop is called for the first time with no arguments, you will be sent directly to this page:

![ConfigPage](https://github.com/raneamri/gotop/blob/main/img/Screenshot%202023-06-23%20at%2017.50.50.png)

To proceed, enter a valid DSN and its DBMS*. Naming the connection is highly recommended if you plan and using multiple connections.

*As of Fri 23rd June 2023, the only DBMS supported is MySQL.

### Processlist
![Processlist](https://github.com/raneamri/gotop/blob/main/img/Screenshot%202023-06-23%20at%2017.51.34.png)

Filtering for processlist coming in July.

### InnoDB Dashboard
![InnoDBDashboard](https://github.com/raneamri/gotop/blob/main/img/Screenshot%202023-06-23%20at%2017.51.54.png)

### Memory Dashboard
![MemoryDashboard](https://github.com/raneamri/gotop/blob/main/img/Screenshot%202023-06-23%20at%2017.52.09.png)

### Error Log
![ErrorLog]

### Lock Log
![LockLog]