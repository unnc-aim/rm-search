#!/bin/bash

BINARY_LOG_FILE="binlog.000137"
START_DATETIME="2025-03-17 00:00:00"
END_DATETIME="2025-03-18 23:00:00"
DATABASE_NAME="rm_search"
TABLE_NAME="search_log"
USER="root"
PASSWORD="123456"
RECOVER_FILE="recover.sql"

mysqlbinlog --database="$DATABASE_NAME" --start-datetime="$START_DATETIME" --stop-datetime="$END_DATETIME" $BINARY_LOG_FILE > $RECOVER_FILE
mysql --host=127.0.0.1 --port=3306 -u $USER -p$PASSWORD $DATABASE_NAME < $RECOVER_FILE
