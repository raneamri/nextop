#!/bin/bash

mysqlslap --user=root --password=961544 --host=localhost --create-schema=testdb --concurrency=10 --iterations=10000 --query=/Users/imraneamri/Documents/nextop/benchmarking/suite1.sql