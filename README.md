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
    ./gotop <dbms> <user> <pass> <host;default:127.0.0.1> <port;default:3306> <db-name;default:none>
```
where the last three arguments are non-ordinal.

Valid examples include:
```bash
    ./gotop mysql root mypass mydatabase
    ./gotop MySQL user pass 3306 127.0.0.1 databasename
```

Upon successful TCP connection, instance will be written in config for easy access in the future.

## Interface
Coming soon.
