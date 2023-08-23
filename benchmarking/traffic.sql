-- commands
--mysqlslap --user=root --password=961544 --host=localhost --create-schema=testdb --concurrency=10 --iterations=10000 --query=/Users/imraneamri/Documents/nextop/benchmarking/traffic.sql

-- queries
INSERT INTO testtable (id, data) VALUES (FLOOR(RAND() * 10000000), 'data_a');

--START TRANSACTION;
--SELECT * FROM testtable WHERE id = 1;
--UPDATE testtable SET data = 'updated' WHERE id = 1;
--COMMIT;