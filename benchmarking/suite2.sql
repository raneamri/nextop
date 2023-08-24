-- lock simulation test

USE testdb;
START TRANSACTION;
SELECT * FROM testtable WHERE some_column = 'value' FOR UPDATE;
COMMIT;