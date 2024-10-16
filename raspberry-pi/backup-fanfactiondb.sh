#!/bin/bash

DB_PATH="/home/tmffuser/fanfaction.db"
BACKUP_PATH="/home/tmffuser/backups/fanfaction_$(date +\%Y\%m\%d).db"

sqlite3 $DB_PATH ".backup $BACKUP_PATH"

find /home/tmffuser/backups/ -name "fanfaction*.db" -type f -mtime +7 -exec rm {} \;