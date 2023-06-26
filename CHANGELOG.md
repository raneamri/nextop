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

## Mon 19th Jun 2023:
- Added data to Processlist page.
- Added dynamic display. (dynamic.go)
- Added connection authentification.
- Small optimisiations within ui package.
- Began writing keybinds and help page.
- Began writing config page and taking input.
- Set groundwork for all pages.

## Tue 20th Jun 2023:
- Added InnoDB data retrieving and parsing.
- Changed approach to data parsing.
- Flipped processlist (it was upside down).
- Fixed potential segfault.
- Implemented line chart for queries.

## Wed 21st Jun 2023:
- Implementing InnoDB dashboard page.
- Changed request method for ease of modulation.
- Added long_queries.go
- Added byte (int) to MiB (string) conversion.
- Retrieved most data needed for InnoDB page.
- Added barchart for selects etc.

## Thu 22 Jun 2023:
- Set layout for Error Log & Memory Dashboard.
- Fixed processlist. Now shows true queries and ignores illegal chars. aswell as resets on tick.
- Changed db-name to conn-name.
- Added active connections slice.
- Fixed some aspect-ratios.
- Added error logging to configs as well as fixed a potential segfault.
- Fixed long lasting instance duplication error.
- Configs interface improved.

## Fri 23 Jun 2023:
- Fixed my unconventional GoLang practices.
- Much better connection pool, using three queues.
- Added in-interface connection cycling.
- Formatted some of the data better.

## Mon 26 Jun 2023:
- Added missing queries for InnoDB Dashboard.
- Better formatted Memory Dashboard.
- Added error log.
- Added specific refresh rate for error log.
- Added include and exclude filters to error log.
- Added linechart to error log.
- Added settings display in config tab.