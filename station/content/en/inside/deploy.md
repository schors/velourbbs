# Deploy

A new version of the station comes out on a tag. A maintainer pushes a tag such as `v0.1.0`
on a commit of `main`. GitHub Actions does the rest.

## Build

The job checks that the tag is on a commit of `main`. Then it runs `go vet` and the tests
and builds Velour for Linux, macOS and Windows.
The programs and their SHA-256 sums go to a GitHub Release.

## Delivery

The job sends one archive to the host over SSH: the program, the directory `station/`,
the systemd units and the install script. The host address and the SSH key are
GitHub secrets. The job log is public, but it has no secrets.

## Install

The script on the host first runs `velourbbs check` on the new station.
If the check finds an error, the deploy stops and the earlier version keeps running.
Then the script points the link `current` to the new version and restarts the service.

The data file is outside the versions and stays after each deploy.
The screen "About the station" shows the tag in the row "Velour version".

## No lost connections

systemd holds the port, not Velour. During a restart the port stays open:
new connections wait in a queue, and the new version accepts them.
The open sessions show the restart screen and close.

## Rollback

To bring back an earlier version, a maintainer re-runs its run in GitHub Actions.
GitHub allows this for 30 days.
