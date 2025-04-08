# Loki log collector & database

Loki is a log collector & storage engine, consisting of multiple different processes.
To make deployment easy, every process is contained in the same binary and can be enabled individually using the `-target=` flag. Targets may be listed using the `-list-targets` flag.

## Simple scalable deployment

See https://grafana.com/docs/loki/latest/get-started/deployment-modes/#simple-scalable.

Loki’s simple scalable deployment mode separates execution paths into read, write, and backend targets. These targets can be scaled independently, letting you customize your Loki deployment to meet your business needs for log ingestion and log query so that your infrastructure costs better match how you use Loki.

The simple scalable deployment mode can scale up to a few TBs of logs per day, however if you go much beyond this, the microservices mode will be a better choice for most users.

The three execution paths in simple scalable mode are each activated by appending the following arguments to Loki on startup:

    -target=write - The write target is stateful and is controlled by a Kubernetes StatefulSet. It contains the following components:
        Distributor
        Ingester
    -target=read - The read target is stateless and can be run as a Kubernetes Deployment that can be scaled automatically (Note that in the official helm chart it is currently deployed as a stateful set). It contains the following components:
        Query Frontend
        Querier
    -target=backend - The backend target is stateful, and is controlled by a Kubernetes StatefulSet. Contains the following components:
        Compactor
        Index Gateway
        Query Scheduler
        Ruler
        Bloom Planner (experimental)
        Bloom Builder (experimental)
        Bloom Gateway (experimental)

The simple scalable deployment mode requires a reverse proxy to be deployed in front of Loki, to direct client API requests to either the read or write nodes. The Loki Helm chart includes a default reverse proxy configuration, using Nginx.
