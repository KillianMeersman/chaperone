FROM prom/prometheus:main

COPY prometheus.yaml /etc/prometheus/prometheus.yml

VOLUME [ "/prometheus" ]
CMD [ "--config.file=/etc/prometheus/prometheus.yml", "--storage.tsdb.path=/prometheus", "--web.enable-otlp-receiver", "--web.enable-remote-write-receiver"]
