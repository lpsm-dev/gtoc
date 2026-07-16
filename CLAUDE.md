# CLAUDE.md

This file guides Claude Code (and other Claude-based agents) when working in
this repository. The full, tool-agnostic contribution guidance lives in
[AGENTS.md](AGENTS.md); this file imports it so both stay in sync.

@AGENTS.md

## Claude-specific notes

- Treat the rules in `AGENTS.md` as hard constraints, not suggestions — in
  particular the cyclomatic-complexity ≤ 10 guardrail and the GitHub-compatible
  anchor algorithm.
- Prefer editing `internal/generator` behavior behind its existing tests: change
  a test to describe the new behavior first, then make it pass.
- After any change that affects TOC output, dogfood it: build the binary and run
  `gtoc generate README.md` to verify the real result, not just unit tests.
- When touching dependencies, remember Go v2+ modules may move their import path
  (e.g. `charm.land/glamour/v2`, `charm.land/log/v2`); update imports and run a
  full `go build ./...` before concluding.
