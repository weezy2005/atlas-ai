import time
from collections.abc import Awaitable, Callable

from fastapi import Request, Response
from prometheus_client import Counter, Histogram
from starlette.routing import Match

REQUESTS = Counter("http_requests_total", "HTTP requests", ["service", "method", "route", "status"])
LATENCY = Histogram(
    "http_request_duration_seconds",
    "HTTP request latency",
    ["service", "method", "route"],
    buckets=(0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10),
)

Middleware = Callable[[Request, Callable[[Request], Awaitable[Response]]], Awaitable[Response]]


def _route_template(request: Request) -> str:
    # Label by route template, not raw path, to keep label cardinality bounded.
    for route in request.app.router.routes:
        match, _ = route.matches(request.scope)
        if match is Match.FULL:
            return str(getattr(route, "path", "unknown"))
    return "unmatched"


def metrics_middleware(service: str) -> Middleware:
    async def middleware(
        request: Request, call_next: Callable[[Request], Awaitable[Response]]
    ) -> Response:
        start = time.perf_counter()
        route = _route_template(request)
        status = 500
        try:
            response = await call_next(request)
            status = response.status_code
            return response
        finally:
            if route != "/metrics":
                LATENCY.labels(service, request.method, route).observe(time.perf_counter() - start)
                REQUESTS.labels(service, request.method, route, str(status)).inc()

    return middleware
