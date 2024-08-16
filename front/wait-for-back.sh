#!/bin/sh
set -e

# Wait for the back service to be available
until curl -s http://back:8080/ | grep -q 'Welcome to LP'; do
  >&2 echo "Back service is unavailable - sleeping"
  sleep 5
done

>&2 echo "Back service is up - starting front"
exec "$@"