#!/usr/bin/env python3
"""LOCAL DEVELOPMENT ONLY. Explicitly advance one order via the existing staff API."""
import argparse
import getpass
import http.client
import json
import os
from pathlib import Path
import sys

STATUSES = ("confirmed", "preparing", "ready_for_delivery", "out_for_delivery", "delivered")
ENV_KEYS = ("APP_ENV", "GO_ENV", "ENVIRONMENT", "NODE_ENV")
LOCAL_ENVS = ("local", "dev", "development")


def check_local(env, env_file):
    if env.get("FOODHUB_DEMO_STATUS") != "1" or env.get("APP_ENV") not in LOCAL_ENVS:
        raise ValueError("Disabled: explicitly set APP_ENV=development FOODHUB_DEMO_STATUS=1 for local use.")
    for key in ENV_KEYS:
        for value in (env.get(key), env_file.get(key)):
            if value and value.lower() not in LOCAL_ENVS:
                raise ValueError("Disabled: a non-development environment is configured.")


def file_environment(path):
    # Only inspect environment labels. Never source .env or display its contents.
    values = {}
    if path.exists():
        for line in path.read_text().splitlines():
            key, separator, value = line.strip().removeprefix("export ").partition("=")
            if separator and key.strip() in ENV_KEYS:
                values[key.strip()] = value.strip().strip("\"'")
    return values


def request(method, path, token, body=None, port=8080):
    # A fixed loopback host and http.client prevent remote targets and redirects.
    connection = http.client.HTTPConnection("127.0.0.1", port, timeout=10)
    try:
        headers = {"Authorization": "Bearer " + token}
        if body is not None:
            headers["Content-Type"] = "application/json"
        connection.request(method, "/api" + path, json.dumps(body) if body is not None else None, headers)
        response = connection.getresponse()
        payload = json.loads(response.read())
        if response.status >= 300 or not payload.get("success"):
            raise ValueError("API rejected request (HTTP %s): %s" % (response.status, payload.get("message", "request failed")))
        return payload["data"]
    finally:
        connection.close()


def advance(order_id, status, token, port=8080):
    if order_id < 1 or status not in STATUSES:
        raise ValueError("Choose a positive order ID and a supported forward status.")
    if not token.strip() or any(char.isspace() for char in token):
        raise ValueError("A staff JWT is required.")
    user = request("GET", "/auth/me", token, port=port)
    if user.get("role") not in ("manager", "admin"):
        raise ValueError("Only an existing manager or admin can advance orders.")
    # The server enforces restaurant ownership, next-step transitions and races.
    return request("PATCH", "/orders/%d/status" % order_id, token, {"status": status}, port=port)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("order_id", type=int)
    parser.add_argument("status", choices=STATUSES)
    parser.add_argument("--port", type=int, default=8080)
    args = parser.parse_args()
    try:
        check_local(os.environ, file_environment(Path(__file__).resolve().parents[2] / ".env"))
        if not 1 <= args.port <= 65535:
            raise ValueError("Invalid local port.")
        print("LOCAL DEMO: this will change one order's status using existing staff permissions.")
        token = getpass.getpass("Existing manager/admin JWT (hidden): ")
        if input("Type ADVANCE to perform this status change: ") != "ADVANCE":
            raise ValueError("No status change requested.")
        result = advance(args.order_id, args.status, token, args.port)
        print("Order %s: %s" % (result["order_number"], result["status"]))
    except (ValueError, OSError, http.client.HTTPException):
        # Do not print headers, tokens or arbitrary server data on failure.
        print("Stopped: check local-development settings, staff authentication, ownership and the allowed next status.", file=sys.stderr)
        return 1
    return 0


if __name__ == "__main__":
    sys.exit(main())
