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
# Help

# Processlist

# InnoDB Dashboard

# Memory Dashboard

# Error Log

# Lock Log

# Configs