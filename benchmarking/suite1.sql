-- classic test
-- aims to simulate simple traffic

-- query
USE testdb;
INSERT INTO testtable (id, data) VALUES (FLOOR(RAND() * 10000000), '_data');