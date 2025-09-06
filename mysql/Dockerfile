FROM mysql:8.0

USER root

RUN microdnf install -y tzdata

RUN mysql_tzinfo_to_sql /usr/share/zoneinfo > /docker-entrypoint-initdb.d/load_timezones.sql

USER mysql