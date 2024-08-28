#!/bin/bash

pg_hostname="eric-oss-byos-postgres"
pg_username="restsim"
pg_password="restsim"
tablename="modb"
#dumpfile=""
for TAB in ${tablename}; do
PGPASSWORD="$pg_password" pg_dump -h "$pg_hostname" -U "$pg_username" --table "$TAB" > "$TAB".sql
done
echo "Dumps created" >exist.txt
