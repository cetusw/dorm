#!/bin/bash
set -euo pipefail

FRONTEND_DIR="frontend"

if [ ! -d "$FRONTEND_DIR" ]; then
  echo "Frontend directory not found: $FRONTEND_DIR"
  exit 1
fi

COMMAND="${1:-}"

case "$COMMAND" in
  install)
    echo "Installing frontend dependencies..."
    cd "$FRONTEND_DIR"
    npm install
    ;;

  build)
    echo "Building frontend..."
    cd "$FRONTEND_DIR"
    npm run build
    ;;

  dev)
    echo "Starting frontend dev server..."
    cd "$FRONTEND_DIR"
    npm run dev -- --host 0.0.0.0
    ;;

  *)
    echo "Usage:"
    echo "  ./bin/frontend.sh install"
    echo "  ./bin/frontend.sh build"
    echo "  ./bin/frontend.sh dev"
    exit 1
    ;;
esac