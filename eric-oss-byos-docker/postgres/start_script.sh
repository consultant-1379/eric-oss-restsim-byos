#!/bin/bash/

CHECK_FILE=/var/lib/postgresql/data/check.txt

if [ -f "$CHECK_FILE" ]; then

echo "file exists"

else

for i in *.sql; do

psql -h localhost -U restsim -d restsim < $i

done

touch /var/lib/postgresql/data/check.txt

fi
