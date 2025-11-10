#!/usr/bin/env python3
import argparse
import subprocess
import signal
import time
import sys

def parse_time(s: str) -> float:
    """Parse durations like 5m, 10s, 2h into seconds."""
    s = s.strip().lower()
    if s.endswith("s"):
        return float(s[:-1])
    elif s.endswith("m"):
        return float(s[:-1]) * 60
    elif s.endswith("h"):
        return float(s[:-1]) * 3600
    return float(s)

def run_with_sigint(cmd, after, timeout, use_sudo=False):
    if use_sudo and cmd[0] != "sudo":
        cmd = ["sudo"] + cmd

    print(f"[+] Running: {' '.join(cmd)}")
    proc = subprocess.Popen(cmd)

    try:
        time.sleep(after)
        print(f"[+] Sending SIGINT to PID {proc.pid} after {after}s")
        proc.send_signal(signal.SIGINT)

        try:
            proc.wait(timeout=timeout)
            print(f"[+] Process exited with code {proc.returncode}")
            sys.exit(proc.returncode)
        except subprocess.TimeoutExpired:
            print(f"[!] Process did not exit after {timeout}s, killing it...")
            proc.kill()
            proc.wait()
            sys.exit(proc.returncode)
    finally:
        if proc.poll() is None:
            proc.kill()
            proc.wait()

def main():
    # Everything before '--' are args for runsigint; everything after is the command.
    if "--" not in sys.argv:
        print("Usage: runsigint --after 2m --timeout 5m [--sudo] -- <command> [args...]")
        sys.exit(1)

    idx = sys.argv.index("--")
    rs_args = sys.argv[1:idx]
    cmd = sys.argv[idx + 1:]

    parser = argparse.ArgumentParser(
        description="Run a command, send SIGINT after N seconds, and wait for it to exit."
    )
    parser.add_argument("--after", required=True, help="When to send SIGINT (e.g., 2m, 30s, 1h)")
    parser.add_argument("--timeout", required=True, help="Max wait after SIGINT (e.g., 5m)")
    parser.add_argument("--sudo", action="store_true", help="Run command with sudo")
    args = parser.parse_args(rs_args)

    after = parse_time(args.after)
    timeout = parse_time(args.timeout)

    run_with_sigint(cmd, after, timeout, args.sudo)

if __name__ == "__main__":
    main()
