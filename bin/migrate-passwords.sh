#!/usr/bin/env bash
set -euo pipefail
umask 077
ROOT=$(cd "$(dirname "$0")/.." && pwd -P)
DRY_RUN=''
if [[ "${1:-}" == --dry-run ]]; then DRY_RUN=$1; shift; fi
[[ $# == 1 ]] || { echo "Usage: $0 [--dry-run] /absolute/path/passwords.txt" >&2; exit 1; }
INPUT=$1
[[ "$INPUT" == /* && -f "$INPUT" && ! -L "$INPUT" && -O "$INPUT" && $(stat -c '%a %h' "$INPUT") == '600 1' ]] || { echo 'Input must be an owned mode-600 regular file without links' >&2; exit 1; }
[[ $(stat -c %s "$INPUT") -le 1048576 ]] || { echo 'Credential input is too large' >&2; exit 1; }
exec bash "$ROOT/bin/release.sh" "$ROOT" migrate-passwords "$DRY_RUN" < "$INPUT"
