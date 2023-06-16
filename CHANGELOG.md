# PRE-RELEASE PATCHES

## Fri 9th Jun 2023:
- Program can establish connection with MySQL database.
- Receives configurations as terminal args & prompted input.
- (Unfinished) partial sync with .conf.

## Mon 12th Jun 2023:
- Re-packaged into subdirs.
- Default port equalling 0 fixed.
- Hotfixed segfault in no arg. startup mode.
- SQL drivers expire unexpectedly by when they reach InitDashboard() (Unresolved)
- Added Makefile.

## Tue 13th Jun 2023:
- Bugfixing.

## Wed 14th Jun 2023:
- More bugfixing, driver error patched.
- State machine implementation.
- QOL.
- Lock management system implementation started.
- Repo set to private.
- Config file remanaged.

## Thu 15th Jun 2023:
- Simplified i/o.
- Simplified data struct heavily.
- Connection pooling implementation started.
- Beginning UI implementation.
- Processlist & lock acquirement done. (to parse still)

## Fri 16th Jun 2023:
- Formatting processlist query output.
- Formatting ps to us & ms.
- Displaying processlist data.
- Layout for processlist view done.

# OFFICIAL PATCHES