#!/bin/sh
set -e

echo "Migrating the database..."
migrate -path=/migrations -database "$DATABASE_URL" up

echo "Starting the application..."
exec ./app