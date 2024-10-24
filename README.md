# Chaperone

[Chaperone](https://en.wikipedia.org/wiki/Chaperone_(social)) is a forward HTTP proxy that does caching & rate-limiting. It's meant to sit between your workloads and external servers, keeping the amount of simultaneous connections in check and caching responses where possible. This prevents you from overloading servers, getting rate-limited or even IP-banned. It also allows you to keep your code relatively simple: Sending of requests without having to coordinate or consider various HTTP caching or rate-limiting semantics.

> [!IMPORTANT]
> Chaperone does not support the CONNECT verb and can thus not act as a https proxy.
> This is because https proxies act as TCP relays and cannot see the HTTP request or response,
> leaving them unable to cache or rate-limit based on url.
>
> All outgoing requests must be made to http addresses. Set the X-Upgrade-HTTPS header to 'true'
> to make Chaperone upgrade your address to https before sending off the request.

## Configuration
Chaperone takes a configuration file, located at $CONFIGFILE (default: ./chaperone.yaml), where you can specify rate limits & caching overrides. It takes the following format:

Chaperone will always listen on $PORT (default 8080).

```yaml
rate_limits:
  # Limit the number of GET requests to this url to 1000/minute.
  - url: https://example.com
    method: GET
    wait_duration: 0.06s  # 1000 requests/minute
  # Limit the number of POST requests to this url to 100/minute.
  - url: https://example.com/test
    method: POST
    wait_duration: 0.6s  # 100 requests/minute

cache_overrides:
   # Force a cache of at least 10m to 1h on all urls starting with `https://example.com`
   # Responses without cache headers are given a caching ttl of 1m.
  - url: https://example.com
    min_ttl: 10m
    max_ttl: 1h
    default_ttl: 10m
    # Force a cache of at least 1m to 10m on all urls starting with `https://example.com/test`
    # If applicable, this will take precedence over the override above.
  - url: https://example.com/test
    min_ttl: 1m
    max_ttl: 10m
    default_ttl: 1m
```

## Implementation
### Python requests
```python
import requests

CHAPERONE_URL = 'http://127.0.0.1:8080'
proxies = {
  'http': CHAPERONE_URL,
}

response = requests.get("http://example.com", proxies=proxies, headers={"X-Upgrade-HTTPS": "true"})
```