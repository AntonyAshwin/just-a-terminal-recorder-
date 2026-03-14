# Just a Terminal Recorder (REC)

REC is a cross-platform CLI tool (macOS/Linux first) that records your last shell commands and saves them as reusable executable scripts.

## Features

- `rec record [n] <name>`: record the last `n` history lines (default `n=1`) into `~/.jtr/bin/<name>`
- `rec start`: start a recording session
- `rec stop <name>`: stop active session and save captured commands as `<name>`
- `rec run <name>`: execute a stored recording
- `rec list`: list available recordings
- `rec delete <name>`: delete a recording
- `rec edit <name>`: open a recording in `$VISUAL` or `$EDITOR`
- `rec inspect <name>`: print stored commands without executing
- `rec init [zsh|bash]`: print shell integration so recordings can change your current shell state
- `rec setup`: configure shell profile automatically for path-aware playback
- `rec path`: print storage directory
- `rec version`: show version

## Project layout

```
jtr/
  main.go
  go.mod
  cmd/
    root.go
    record.go
    run.go
    list.go
    delete.go
    edit.go
    path.go
    version.go
  internal/
    history/
      reader.go
    recorder/
      recorder.go
    storage/
      storage.go
    executor/
      executor.go
    shell/
      detect.go
```

## Build

```bash
go mod tidy
go build -o rec
```

This creates the `rec` binary in the project root.

## Run

```bash
./rec version
./rec path
```

## Install

Option 1 (global Go install):

```bash
go install .
```

Option 2 (copy built binary manually):

```bash
cp ./rec /usr/local/bin/rec
```

## One-time setup (recommended)

Run a single command after install:

```bash
rec setup
```

This updates your shell profile (`~/.zshrc` or `~/.bashrc`) with:

- `~/.jtr/bin` added to `PATH`
- shell integration via `rec init`

Then reload your shell:

```bash
source ~/.zshrc
```

Use `~/.bashrc` if you are on bash.

With setup enabled, `rec record`, `rec stop`, and `rec delete` automatically refresh recording functions in your current shell. You do not need to run `rec_load_recordings` manually each time.

## Manual setup

If you prefer not to use `rec setup`, you can configure it manually.

### Add recordings directory to PATH

REC stores executable recordings in `~/.jtr/bin`.

Add this once to your shell profile:

```bash
export PATH="$HOME/.jtr/bin:$PATH"
```

## Enable path-aware playback in your current shell

Executable scripts run in a child process, so `cd` in a recording will not change your active terminal directory by default.

Enable shell integration once in your profile:

zsh (`~/.zshrc`):

```bash
eval "$(rec init zsh)"
```

bash (`~/.bashrc`):

```bash
eval "$(rec init bash)"
```

After reloading your shell, recordings are loaded as shell functions. Running a recording like `desktop` can then change your current directory.

## Usage examples

Record three recent commands into `deploy`:

```bash
rec record 3 deploy
```

Session-based recording:

```bash
rec start
git add .
git commit -m "fix bug"
git push
rec stop deploy
```

By default, `stop` saves path context by prepending the session's original start directory,
so recordings replay from the same base location.

To create a non-pathable script (ignore path context and `cd` commands):

```bash
rec stop deploy --ignore-path
```

If `stop` says no commands were captured yet, the session stays active so you can retry.
On zsh, this helps make command history immediately available:

```bash
setopt INC_APPEND_HISTORY SHARE_HISTORY
```

Record one recent command (default count):

```bash
rec record cleanup
```

Replay later:

```bash
deploy
```

Inspect without running:

```bash
rec inspect deploy
```

## Record flags

- `--include-path`: keep path-changing commands (`cd`)
- `--ignore-path`: remove path-changing commands (`cd`)
- `--ip`: alias for `--ignore-path`
- `--absolute-paths`: convert relative `cd` paths to absolute paths
- `--dry-run`: print what would be recorded without saving

Examples:

```bash
rec record 5 deploy --ignore-path
rec record deploy --include-path
rec record 5 deploy --dry-run
```

## Notes

- Shell history support: `zsh` and `bash`
- `rec setup` modifies your selected shell profile idempotently
- Recording names are restricted to letters, numbers, `.`, `_`, and `-`
- When a recording name already exists, REC asks for overwrite confirmation
- `start/stop` filters out `rec` commands so control commands are not recorded
# just-a-terminal-recorder-
