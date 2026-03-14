# vibe.md

## Purpose

This file is a quick handoff for another Copilot/agent working on this repository.

Project goal: record shell commands and save them as reusable scripts/functions that can be replayed later.
User-facing command: `rec`.

## Current command surface

- `rec record [n] <name>`
- `rec start`
- `rec stop <name>`
- `rec run <name>`
- `rec list`
- `rec delete <name>`
- `rec edit <name>`
- `rec inspect <name>`
- `rec init [zsh|bash]`
- `rec setup`
- `rec path`
- `rec version`

## Filesystem structure

```text
.
├── go.mod
├── main.go
├── README.md
├── vibe.md
├── cmd/
│   ├── delete.go
│   ├── edit.go
│   ├── init.go
│   ├── inspect.go
│   ├── list.go
│   ├── path.go
│   ├── record.go
│   ├── root.go
│   ├── run.go
│   ├── setup.go
│   ├── start.go
│   ├── stop.go
│   └── version.go
├── internal/
│   ├── executor/
│   │   └── executor.go
│   ├── history/
│   │   └── reader.go
│   ├── recorder/
│   │   ├── recorder.go
│   │   └── recorder_test.go
│   ├── shell/
│   │   └── detect.go
│   └── storage/
│       └── storage.go
├── jtr
└── rec
```

## High-level architecture

- CLI layer: `cmd/` (Cobra-style command handlers and flags).
- Domain logic:
  - history capture in `internal/history/`
  - recording/session orchestration in `internal/recorder/`
  - script execution in `internal/executor/`
  - shell detection and integration in `internal/shell/`
  - storage paths and persistence in `internal/storage/`
- Entrypoint: `main.go` wires root command and subcommands.

## Runtime/storage behavior

- Recordings are stored under `~/.jtr/bin`.
- Shell setup can be automated by `rec setup`.
- `rec init` outputs shell integration snippets for `zsh` and `bash`.
- Path-aware replay depends on shell integration because plain script execution runs in a child process.

## Build and run

```bash
go mod tidy
go build -o rec
./rec version
./rec path
```

Install globally (manual copy):

```bash
cp ./rec /usr/local/bin/rec
```

## Important implementation notes

- Internal Go imports still use module path `jtr/...` (example: `jtr/internal/recorder`).
- This is expected and separate from the user-facing CLI command name (`rec`).
- A known source of confusion is building only `jtr` while running `rec`; prefer explicit output with `go build -o rec`.

## If you continue work here

- Keep user-facing docs/examples on `rec`.
- Avoid changing storage path semantics (`~/.jtr`) unless migration behavior is planned.
- When editing command docs, make sure `README.md` and this file stay aligned.
