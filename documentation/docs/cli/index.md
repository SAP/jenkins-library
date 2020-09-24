# Project "Piper" CLI

The CLI is built using the go programming language and thus is distributed in a single binary file for Linux.

The latest released version can be downloaded via `wget https://github.com/SAP/jenkins-library/releases/latest/download/piper`.

Specific versions an be downloaded from the [GitHub releases](https://github.com/SAP/jenkins-library/releases) page.

Once available in `$PATH`, it is ready to use.

To verify the version you got, run `piper version`.
To read the online help, run `piper help`.

!!! hint "Use the shell completion"
    For the purpose of interactive usage on the command line, we recommend to setup shell completion scripts.
    Run `piper completion --help` for information on how to set it up for your shell.
    This might need to be updated from time to time to reflect new commands added to piper.

!!! note "Linux only (as of now)"
    Since this is a binary compiled for Linux systems, you won't be able to use it on macOS or Windows systems.
    You might try running it inside Docker on those systems.

If you're interested in using it with GitHub Actions, see [the Project "Piper" Action](https://github.com/SAP/project-piper-action) which makes the tool more convinient to use.
