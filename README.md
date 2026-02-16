# machealth

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Platform: macOS](https://img.shields.io/badge/Platform-macOS-lightgrey.svg)](https://github.com/lu-zhengda/machealth)
[![Homebrew](https://img.shields.io/badge/Homebrew-lu--zhengda/tap-orange.svg)](https://github.com/lu-zhengda/homebrew-tap)

macOS system health checker for AI agents — a single-call health assessment across 8 subsystems with JSON output.

## Install

```bash
brew tap lu-zhengda/tap
brew install machealth
```

## Usage

```
$ machealth
{
  "timestamp": "2026-02-16T10:30:00Z",
  "score": { "status": "green", "value": 95, "reasons": [] },
  "cpu": { "status": "green", "load_avg_1m": 1.42, ... },
  "memory": { "status": "green", "pressure_percent": 68, ... },
  "disk": { "status": "green", "available_gb": 120.5, ... },
  ...
}

$ machealth --human
[OK] System Health: GREEN (score: 95/100)

  [OK] CPU            Load: 1.42/1.68/1.75 (0.18 per core, 8 cores)
  [OK] Memory         Free: 68%, Swap: 0.0/0.0 MB
  [OK] Disk           Available: 120.5/228.3 GB (47.2% used)
  [OK] Thermal        CPU speed limit: 100%
  [OK] iCloud         Caught up
  [OK] Battery        78%, AC Power, charging
  [OK] Time Machine   Idle
  [OK] Network        Reachable via en0 (192.168.1.10)

$ machealth diagnose --human
[!!] System Health: YELLOW (score: 72/100)

  [OK] CPU            Load: 1.42/1.68/1.75 (0.18 per core, 8 cores)
  [!!] Memory         Free: 22%, Swap: 1024.0/2048.0 MB
  [OK] Disk           Available: 120.5/228.3 GB (47.2% used)
  ...

--- Diagnoses (1) ---

1. [YELLOW] memory: Memory pressure elevated
   Detail: Free memory at 22% with 1024.0 MB swap in use
   Action: Close unused applications to free memory
```

## Commands

| Command | Description | Example |
|---------|-------------|---------|
| `check` | One-shot health check (default) | `machealth` or `machealth check` |
| `watch` | Continuously monitor (JSON Lines) | `machealth watch --interval 10s` |
| `watch --human` | Refreshing terminal display | `machealth watch --human` |
| `diagnose` | Health check with actionable diagnoses | `machealth diagnose --human` |
| `--human` | Human-readable output instead of JSON | `machealth --human` |

### Exit Codes

| Code | Meaning | Score |
|------|---------|-------|
| `0` | Healthy (green) | 80-100 |
| `1` | Degraded (yellow) | 50-79 |
| `2` | Critical (red) | 0-49 |

## Subsystems

| Subsystem | Weight | What It Checks |
|-----------|--------|----------------|
| CPU | 20% | Load averages, per-core load |
| Memory | 25% | Memory pressure, swap usage |
| Thermal | 20% | CPU speed limit, throttling |
| Disk | 15% | Available space, usage percent |
| Battery | 10% | Charge level, power source, health |
| iCloud | 5% | Sync status |
| Network | 5% | Reachability, active interface |
| Time Machine | 0% | Backup state (degrades status, not score) |

## Diagnostic Workflow

1. `machealth` — quick health check, exit code tells you the status
2. `machealth diagnose` — detailed diagnoses with suggested actions
3. `machealth watch` — continuous monitoring as JSON Lines for piping
4. `machealth watch --human` — live terminal dashboard

## Claude Code

Available as a skill in the [macos-toolkit](https://github.com/lu-zhengda/macos-toolkit) Claude Code plugin. Ask Claude "check system health" or "diagnose my Mac" and it runs machealth automatically.

## License

[MIT](LICENSE)
