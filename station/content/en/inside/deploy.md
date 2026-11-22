# Deploy

## What it does

A new version of the station comes out by a tag. A maintainer puts an annotated tag
such as `v0.1.0` on a commit of `main`. Then GitHub Actions does everything.

## How

### Build

1. The job checks the tag: its form and that the commit of the tag is in `main`.
2. It runs `go vet` and the tests.
3. It builds Velour with `CGO_ENABLED=0` for linux/amd64, linux/arm64,
   darwin/arm64 and windows/amd64.
4. `go version -m` checks that each program has this tag.
5. The programs and the file `SHA256SUMS` go to the GitHub Release.
   The release text is the annotation of the tag.

The job uses only `actions/checkout` and `actions/setup-go`.
Each is pinned to a full commit SHA. There are no third-party actions for SSH and releases:
they would get the SSH key.

### Delivery

The job sends one archive to the host over SSH: the program, the directory `station/`,
the systemd units and the install script. The SSH key and the host address are in the secrets
of a GitHub environment. The environment allows only the tags `v*`.

### Install

1. The script runs `velourbbs check` on the new station. The check also reads the data file.
   An error stops the deploy, and the previous version keeps running.
2. It puts the version into `releases/<tag>/` and points the link `current` to it.
3. It restarts the service and after 3 seconds checks that it runs.
4. It keeps the 3 newest versions.

The systemd unit `velourbbs.socket` holds the port, not the program. During a restart,
the port stays open: new connections wait in the queue, and the new version accepts them.
Open sessions show the restart screen and close.

## What it gets

- A version that the screen "About the station" shows in the row "Velour version".
- An open job log: it has the steps and the result, but no secrets.
- A data file that is outside the versions and stays after each deploy.

## Limits

- A rollback is a re-run of the earlier version in GitHub Actions. GitHub allows it for 30 days.
  Later, a revert in `main` and a new tag.
- After a migration of the data file, the earlier version fails the check.
  Then stop `velourbbs.socket` and `velourbbs.service`, restore the copy of the file
  of that version and re-run.
- A person who can push a tag `v*` can also change the job. Only the maintainers
  can write to the repository.
