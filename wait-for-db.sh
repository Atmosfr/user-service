#!/bin/sh
set -e

host="db"
port="5432"

until nc -z "$host" "$port"; do
  echo "Waiting for PostgreSQL at $host:$port..."
  sleep 1
done

echo "PostgreSQL is up — starting application"
exec "$@"
