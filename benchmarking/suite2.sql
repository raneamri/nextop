-- lock simulation test

SET innodb_lock_wait_timeout = 5; -- Set lock wait timeout to 5 seconds
BEGIN;
UPDATE your_table SET your_column = your_column + 1 WHERE id = your_id1;
BEGIN;
UPDATE your_table SET your_column = your_column + 1 WHERE id = your_id1; -- Will wait due to lock