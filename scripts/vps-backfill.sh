#!/usr/bin/env sh
# Run historical analytics backfill inside Docker (same network as Postgres).
#
# On the VPS:
#   chmod +x scripts/vps-backfill.sh
#   ./scripts/vps-backfill.sh                           # dry-run, uses ./cscd.csv
#   ./scripts/vps-backfill.sh /opt/leetbot/cscd.csv -apply
#
# From your laptop (foreground):
#   ssh leetbot 'cd /opt/leetbot && ./scripts/vps-backfill.sh /opt/leetbot/cscd.csv'
#
# Background on VPS (disconnect-safe):
#   ssh leetbot 'cd /opt/leetbot && nohup ./scripts/vps-backfill.sh /opt/leetbot/cscd.csv -apply >> backfill.log 2>&1 & echo $!'
#
set -eu

ROOT=$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)
cd "$ROOT"

if test $# -ge 1 && case "$1" in -*) false ;; *) true ;; esac; then
	case "$1" in
	/*)
		CSV_ABS=$1
		;;
	*)
		CSV_ABS=$ROOT/$1
		;;
	esac
	shift
else
	CSV_ABS=$ROOT/cscd.csv
fi

if ! test -r "$CSV_ABS"; then
	echo "vps-backfill: csv not found or unreadable: $CSV_ABS" >&2
	exit 1
fi

exec docker compose run --rm \
	-v "$CSV_ABS:/data/cscd.csv:ro" \
	leetbot ./backfill -csv /data/cscd.csv -prefix '!' "$@"
