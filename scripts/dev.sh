#!/usr/bin/env bash

# Development script to run both Discord bot and web server
# Used by Air for hot reload. Must tear down all children before exit so only one
# Discord session is connected when Air restarts this script.
#
# Before starting, kills stray bot/server processes that still have this repo as
# cwd (e.g. orphaned go run after a crash or an Air race). That avoids multiple
# gateways on the same token and Discord "interaction already acknowledged" spam.

set -euo pipefail
# Job control: without this, non-interactive runs (Air) may not track `jobs -p`,
# so old `go run` processes survive across rebuilds.
set -m

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)"
cd "$REPO_ROOT"

# Current working directory of a process (Linux /proc or macOS lsof).
pid_cwd() {
	local pid="$1"
	if [[ -r "/proc/${pid}/cwd" ]]; then
		readlink -f "/proc/${pid}/cwd" 2>/dev/null || readlink "/proc/${pid}/cwd" 2>/dev/null || return 1
		return 0
	fi
	lsof -a -p "$pid" -d cwd 2>/dev/null | awk 'NR == 2 { print $NF }'
}

# Normalize path for comparison (best-effort).
resolve_dir() {
	local d="$1"
	(cd "$d" 2>/dev/null && pwd -P) || printf '%s' "$d"
}

# PIDs matching dev-relevant command lines (may include other checkouts; cwd filters).
collect_leetbot_dev_pids() {
	local pat
	for pat in \
		'go run.*cmd/bot' \
		'go run.*cmd/server' \
		'whotypes/leetbot/cmd/bot' \
		'whotypes/leetbot/cmd/server'; do
		pgrep -f "$pat" 2>/dev/null || true
	done | { grep -E '^[0-9]+$' || true; } | sort -u
}

kill_stale_leetbot_dev_processes() {
	local pid cwd resolved any=0
	while read -r pid; do
		[[ -z "$pid" ]] && continue
		[[ "$pid" == "$$" ]] && continue
		if ! kill -0 "$pid" 2>/dev/null; then
			continue
		fi
		cwd="$(pid_cwd "$pid" 2>/dev/null || true)"
		[[ -z "$cwd" ]] && continue
		resolved="$(resolve_dir "$cwd")"
		if [[ "$resolved" != "$REPO_ROOT" ]]; then
			continue
		fi
		any=1
		echo "Stopping stray leetbot dev process (pid ${pid})..."
		kill -TERM "$pid" 2>/dev/null || true
	done < <(collect_leetbot_dev_pids)

	if [[ "$any" -eq 1 ]]; then
		sleep 0.5
		while read -r pid; do
			[[ -z "$pid" ]] && continue
			[[ "$pid" == "$$" ]] && continue
			if ! kill -0 "$pid" 2>/dev/null; then
				continue
			fi
			cwd="$(pid_cwd "$pid" 2>/dev/null || true)"
			[[ -z "$cwd" ]] && continue
			resolved="$(resolve_dir "$cwd")"
			if [[ "$resolved" != "$REPO_ROOT" ]]; then
				continue
			fi
			echo "Force-killing stubborn leetbot dev process (pid ${pid})..."
			kill -KILL "$pid" 2>/dev/null || true
		done < <(collect_leetbot_dev_pids)
	fi
}

graceful_stop_job_pids() {
	local pid
	for pid in $(jobs -p 2>/dev/null); do
		if kill -0 "$pid" 2>/dev/null; then
			kill -TERM "$pid" 2>/dev/null || true
		fi
	done
}

force_kill_job_pids() {
	local pid
	for pid in $(jobs -p 2>/dev/null); do
		kill -KILL "$pid" 2>/dev/null || true
	done
}

wait_for_jobs() {
	local i=0
	while [[ $i -lt 50 ]]; do
		if [[ -z "$(jobs -p 2>/dev/null)" ]]; then
			return 0
		fi
		sleep 0.15
		i=$((i + 1))
	done
	return 1
}

cleanup() {
	trap - SIGINT SIGTERM EXIT
	set +e
	echo "Stopping dev processes (graceful)..."
	graceful_stop_job_pids
	if ! wait_for_jobs; then
		echo "Some jobs did not exit in time; sending SIGKILL..."
		force_kill_job_pids
	fi
	wait 2>/dev/null || true
	# Catch any bot/server children no longer listed as shell jobs (orphans).
	kill_stale_leetbot_dev_processes
	if command -v lsof >/dev/null 2>&1; then
		lsof -ti:8080 2>/dev/null | while read -r p; do kill -TERM "$p" 2>/dev/null || true; done
		lsof -ti:5173 2>/dev/null | while read -r p; do kill -TERM "$p" 2>/dev/null || true; done
	fi
	set -e
	exit 0
}

trap cleanup SIGINT SIGTERM EXIT

kill_stale_leetbot_dev_processes

echo "Starting web build in watch mode..."
(cd web && bun run watch) &

echo "Starting leetbot and web server..."
go run ./cmd/bot &
go run ./cmd/server &

wait
