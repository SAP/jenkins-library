# Git Hooks

From [Git docs](https://git-scm.com/docs/githooks#_description):

> Hooks are programs you can place in a hooks directory to trigger actions at certain points in git’s execution. Hooks that don’t have the executable bit set are ignored.

## Usage

To use the hook, execute this command in the project root directory to link the script into the `.git/hooks` directory:

```sh
ln -s -f ../../.hooks/pre-commit ./.git/hooks/pre-commit
```

Make sure the file is executable:

```sh
chmod +x ./.hooks/pre-commit
```

## Pre-Commit Hook

From [Git docs](https://git-scm.com/docs/githooks#_pre_commit):

> This hook is invoked by git-commit, and can be bypassed with the --no-verify option. It takes no parameters, and is invoked before obtaining the proposed commit log message and making a commit. Exiting with a non-zero status from this script causes the git commit command to abort before creating a commit.

### Content

Executes `go mod tidy` and stages `go.mod` & `go.sum`.

From [Golang docs](https://github.com/golang/go/wiki/Modules):

> `go mod tidy` — Prune any no-longer-needed dependencies from `go.mod` and add any dependencies needed for other combinations of OS, architecture, and build tags (details).
