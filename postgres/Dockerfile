FROM postgres:10.1

COPY *.sql /

COPY eric-oss-restsim-dataset /

COPY start_script.sh /

COPY dataset_dump.sh /

COPY entrypoint.sh /

RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]

EXPOSE 5432

