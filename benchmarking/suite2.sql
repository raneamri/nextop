-- lock simulation test

START TRANSACTION;
SELECT MAX(id) INTO @max_id FROM testtable;
SET @new_id = @max_id + 1;
INSERT INTO testtable (id, data) VALUES (@new_id, 'newdata');
COMMIT;