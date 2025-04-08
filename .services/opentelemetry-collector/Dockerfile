FROM otel/opentelemetry-collector-contrib:0.123.0

COPY ./config.yaml /etc/otelcol-contrib/config.yaml

# OTLP gRPC receiver
EXPOSE 4317
# OTLP http receiver
EXPOSE 4318
