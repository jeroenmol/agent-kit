#!/bin/sh
# Run this only from a normal local Terminal, not from the managed Codex host.
# It checks that a fresh noninteractive Codex session can start. It does not
# test Agent Kit activation, role isolation, permissions, or protected effects.
set -eu

if ! command -v python3 >/dev/null 2>&1; then
  printf '%s\n' 'check-runtime: python3 is required for timeout cleanup' >&2
  exit 2
fi

PROBE_ROOT=$(mktemp -d "${TMPDIR:-/private/tmp}/agent-kit-startup.XXXXXX")
chmod 700 "$PROBE_ROOT"

export PROBE_ROOT
export CODEX_BIN=${CODEX_BIN:-codex}
export PROBE_TIMEOUT_SECONDS=${PROBE_TIMEOUT_SECONDS:-120}

python3 - <<'PY'
import os
import pathlib
import signal
import subprocess
import sys
import threading
import time

probe_root = pathlib.Path(os.environ["PROBE_ROOT"])
codex_bin = os.environ["CODEX_BIN"]
timeout_seconds = int(os.environ["PROBE_TIMEOUT_SECONDS"])
if not 1 <= timeout_seconds <= 120:
    raise ValueError("PROBE_TIMEOUT_SECONDS must be between 1 and 120")
last_message = probe_root / "last-message.txt"
status_file = probe_root / "status.txt"
process = None
stderr_prefix = bytearray()
stderr_thread = None

command = [
    codex_bin,
    "-a", "never",
    "exec",
    "--ephemeral",
    "--ignore-user-config",
    "--strict-config",
    "-c", 'model_reasoning_effort="medium"',
    "--model", "gpt-5.6-terra",
    "--sandbox", "read-only",
    "--skip-git-repo-check",
    "--output-last-message", str(last_message),
    "-C", str(probe_root),
    "Reply with exactly READY. Do not use tools or read or write files.",
]

result = "failed"
exit_code = "unavailable"
diagnostic = "unclassified"

def stop_process_group():
    if process is None:
        return
    # The leader can exit while a child still owns the process group, so do
    # not use process.poll() as evidence that the group is gone.
    try:
        os.killpg(process.pid, signal.SIGTERM)
    except (PermissionError, ProcessLookupError):
        return
    time.sleep(0.1)
    try:
        os.killpg(process.pid, signal.SIGKILL)
    except (PermissionError, ProcessLookupError):
        pass
    if process.poll() is None:
        process.wait()

def drain_stderr(stream):
    while True:
        chunk = stream.read(1024)
        if not chunk:
            return
        remaining = 4096 - len(stderr_prefix)
        if remaining > 0:
            stderr_prefix.extend(chunk[:remaining])

def cancelled(signum, frame):
    raise KeyboardInterrupt

for sig in (signal.SIGHUP, signal.SIGINT, signal.SIGTERM):
    signal.signal(sig, cancelled)

try:
    process = subprocess.Popen(
        command,
        cwd=probe_root,
        stdin=subprocess.DEVNULL,
        stdout=subprocess.DEVNULL,
        stderr=subprocess.PIPE,
        start_new_session=True,
    )
    stderr_thread = threading.Thread(target=drain_stderr, args=(process.stderr,), daemon=True)
    stderr_thread.start()
    try:
        exit_code = str(process.wait(timeout=timeout_seconds))
    except subprocess.TimeoutExpired:
        result = "timed_out"
        diagnostic = "startup_timeout"
        stop_process_group()
    else:
        try:
            response = last_message.read_text(encoding="utf-8")
        except OSError:
            response = ""
        if exit_code == "0" and response.strip() == "READY":
            result = "started"
            diagnostic = "ready_response"
        elif exit_code == "0":
            diagnostic = "unexpected_response"
        else:
            stderr_thread.join(timeout=1)
            error_text = bytes(stderr_prefix).decode("utf-8", errors="replace").lower()
            if "failed to initialize in-process app-server client" in error_text or "operation not permitted" in error_text:
                diagnostic = "initialization_denied"
            elif "authentication" in error_text or "not logged in" in error_text or "auth" in error_text:
                diagnostic = "auth_unavailable"
            elif "unknown configuration" in error_text or "unrecognized" in error_text or "invalid value" in error_text:
                diagnostic = "configuration_rejected"
            elif "model" in error_text:
                diagnostic = "model_unavailable"
            else:
                diagnostic = "command_failed"
except FileNotFoundError:
    result = "codex_not_found"
    diagnostic = "executable_not_found"
except OSError:
    result = "launch_error"
    diagnostic = "process_launch_error"
except KeyboardInterrupt:
    result = "cancelled"
    diagnostic = "cancelled_by_user"
finally:
    stop_process_group()
    if process is not None and process.stderr is not None:
        process.stderr.close()
    if stderr_thread is not None:
        stderr_thread.join(timeout=1)
    last_message.unlink(missing_ok=True)

status_file.write_text(
    "host_startup_check=" + result + "\n"
    "process_exit=" + exit_code + "\n"
    "diagnostic=" + diagnostic + "\n"
    "requested_model=gpt-5.6-terra\n"
    "requested_reasoning_effort=medium\n",
    encoding="utf-8",
)

# The response and the bounded in-memory stderr prefix are intentionally not
# retained. This is the only diagnostic artifact, and it contains no
# environment, configuration, or auth values.
print(status_file.read_text(encoding="utf-8"), end="")
print("sanitized_artifact=" + str(status_file))
sys.exit(0 if result == "started" else 130 if result == "cancelled" else 1)
PY
