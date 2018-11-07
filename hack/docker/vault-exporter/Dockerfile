FROM prom/statsd-exporter:latest

RUN mkdir -p /conf

COPY STATSD.MAPPING-CONFIG /conf/STATSD.MAPPING-CONFIG

ENTRYPOINT ["/bin/statsd_exporter", "--statsd.mapping-config=/conf/STATSD.MAPPING-CONFIG"]