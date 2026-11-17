# Release and deploy

This specification defines the release of `velourbbs` on a version tag and the deploy
of the home BBS to its host.
Related parts: [home-bbs](home-bbs.md), [telnet-term](telnet-term.md), section 10,
[session-fields](session-fields.md), section 3.4, and [setup-guide](setup-guide.md).

## 1. Goal

A maintainer pushes a tag `vX.Y.Z` on a commit of `main`. A GitHub Actions workflow
builds the binaries, creates a GitHub Release and deploys the home BBS to the host.
`bbs.version` on the host shows the tag.

## 2. Files

| Path | Content |
|---|---|
| `.github/workflows/release.yml` | The workflow of sections 3 and 4. |
| `deploy/velourbbs.service` | The service unit of section 5. |
| `deploy/velourbbs.socket` | The socket unit of section 5, with the placeholder `@PORT@`. |
| `deploy/install.sh` | The host script of section 4.5. |

## 3. Release

### 3.1. Trigger

- The workflow runs on a push of a tag that matches `v*`.
- It has one job. The steps of section 3 come before the steps of section 4.
  A failed step stops the job.
- The job checks the tag first. A tag matches
  `^v(0|1)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z.-]+)?$`. Else the job fails.
  Go writes a tag to the build info only for the major version 0 or 1
  ([session-fields](session-fields.md), section 3.4).
- The job fetches `refs/heads/main` into `refs/remotes/origin/main`. The commit of the tag
  must be an ancestor of `refs/remotes/origin/main^{commit}` (`git merge-base --is-ancestor`).
  Else the job fails, with no release and no deploy. The job uses only full ref names:
  a tag named `origin/main` must not replace the branch.
- The job checks the settings of section 4.1 before the build.
- A pre-release tag (`v0.1.0-rc.1`) gives a GitHub pre-release. It also deploys.
- The workflow has the permission `contents: write` and no other permission.
- One concurrency group with `queue: max` holds all runs of the workflow. A run waits
  for the run before it. It does not cancel it. GitHub keeps at most 100 waiting runs.

### 3.2. Build

- The checkout is the tagged commit with the full history and the tags (`fetch-depth: 0`).
- The job deletes each local tag on the commit except the tag of the run
  (`git tag --points-at HEAD`). Go writes the highest tag of the commit to the build info.
  So `v0.1.0-rc.1` and `v0.1.0` on one commit give the version of the run.
  The delete does not change the working tree.
- Go is the version of `go.mod` (`go-version-file`).
- The job runs `go vet ./...` and `go test ./...` before the build.
- It builds `./cmd/velourbbs` with `CGO_ENABLED=0` and `-trimpath` for these platforms:
  linux/amd64, linux/arm64, darwin/arm64 and windows/amd64.
- The output goes to `$RUNNER_TEMP`, outside the checkout. A new file in the checkout
  gives the version `+dirty` ([session-fields](session-fields.md), section 3.4).
- `go version -m` checks each binary. Its `mod` line must have the tag as the version.
  Else the job fails.

### 3.3. GitHub Release

- The assets are `velourbbs-<tag>-<goos>-<goarch>`, with `.exe` for windows,
  and `SHA256SUMS` with the SHA-256 of each binary (the format of `sha256sum`).
- The job creates the release with `gh release create <tag> --verify-tag`,
  with `--prerelease` for a pre-release tag, and the assets.
- The release notes come from `--generate-notes`.
- When the release of the tag exists, the job does not change it. A re-run (section 4.7)
  then builds and deploys again.

## 4. Deploy

### 4.1. Settings

The GitHub environment `home` gives the deploy settings. Its rule "Deployment branches
and tags" allows only the tags `v*`. The job runs in this environment.
A workflow of a branch push cannot get the SSH key.

| Name | Kind | Content |
|---|---|---|
| `DEPLOY_HOST` | variable | The host name or address for SSH. |
| `DEPLOY_USER` | variable | The user on the host. |
| `DEPLOY_SSH_PORT` | variable | The SSH port. Empty gives 22. |
| `DEPLOY_KNOWN_HOSTS` | variable | The `known_hosts` lines of the host keys. |
| `BBS_PORT` | variable | The TCP port of the BBS in the socket unit. |
| `DEPLOY_SSH_KEY` | secret | The private SSH key of the user, with no passphrase. |

- A missing setting, except `DEPLOY_SSH_PORT`, fails the job before the build.
- `BBS_PORT` and `DEPLOY_SSH_PORT` must be decimal numbers from 1 to 65535.
  `DEPLOY_USER` must match `^[a-z_][a-z0-9_-]*$`. `DEPLOY_HOST` must match `^[A-Za-z0-9.:-]+$`
  and must not start with `-`. Else the job fails before the build. A newline in `BBS_PORT`
  would add a line to the socket unit.
- `DEPLOY_KNOWN_HOSTS` is not secret: it holds public keys. For a port other than 22,
  a line starts with `[<host>]:<port>`. `ssh-keyscan -p <port> <host>` gives the lines.
  Compare their fingerprints with the host before the first deploy.

### 4.2. SSH

- The job uses the OpenSSH client of the runner.
- The key file and the `known_hosts` file go to `$RUNNER_TEMP` with mode 600.
- Each `ssh` call has these options: `-o StrictHostKeyChecking=yes`,
  `-o UserKnownHostsFile=<file>`, `-o IdentitiesOnly=yes`, `-o BatchMode=yes`,
  `-i <key file>` and `-p <port>`.
- A changed host key fails the job. The operator checks the host and updates
  `DEPLOY_KNOWN_HOSTS`.

### 4.3. Trust

- A person who can push a tag `v*` can change the workflow in the tagged commit.
  This workflow gets the SSH key: the check of section 3.1 is in the same file.
- Only the maintainers have write access ([flow](../dev-flow/flow.md), section 1).
  A tag push from a fork does not run the workflow of this repository.

### 4.4. Bundle

- The job reads the host architecture with `ssh ... uname -m`.
  `x86_64` gives amd64. `aarch64` gives arm64. Another value fails the job.
- The bundle is one gzip tar stream with these entries:
  - `velourbbs`: the linux binary of the host architecture;
  - `station/`: the home BBS with the symlinks resolved (`tar -h`), as in
    [home-bbs](home-bbs.md), section 2;
  - `units/velourbbs.service` and `units/velourbbs.socket`: the files of `deploy/`.
    `@PORT@` in the socket unit is `BBS_PORT`;
  - `install.sh`: `deploy/install.sh`.
- The job sends the bundle over SSH into `~/velourbbs/incoming/<tag>/` on the host.
  It removes this directory before. Then it runs `sh ~/velourbbs/incoming/<tag>/install.sh <tag>`.

### 4.5. Host

Layout under the home directory of the user:

| Path | Content |
|---|---|
| `~/velourbbs/releases/<tag>/` | The binary `velourbbs` and `station/` of one release. |
| `~/velourbbs/current` | A symlink to `releases/<tag>` of the running release. |
| `~/velourbbs/incoming/<tag>/` | The bundle during a deploy. |
| `~/.config/systemd/user/velourbbs.service`, `velourbbs.socket` | The units of section 5. |

`install.sh <tag>` does these steps. It stops at the first error (`set -eu`).

1. Run `../velourbbs check` in `incoming/<tag>/station/`. An error stops the deploy.
   The running release does not change.
2. Replace `releases/<tag>/` with the binary and `station/` of the bundle.
3. Compare the socket unit of the bundle with the installed one. Copy the units
   to `~/.config/systemd/user/`. Run `systemctl --user daemon-reload`
   and `systemctl --user enable velourbbs.socket velourbbs.service`.
4. Point `current` to `releases/<tag>`: make a new symlink and rename it over `current`.
5. When the socket unit changed, run `systemctl --user restart velourbbs.socket`.
   `Requires=` also restarts the service. Else run `systemctl --user restart velourbbs.service`.
6. Wait 3 s. `systemctl --user is-active velourbbs.service` must give `active`.
7. Remove `incoming/<tag>/`. Remove the releases except the 3 newest by modification time
   and the target of `current`.

- A restart stops the sessions with the screen `shutdown` ([menu-nav](menu-nav.md),
  section 5.2). The socket stays open during the restart: systemd holds it. New connects
  wait in its queue and the new process accepts them.
- `systemctl --user` needs the user manager. Linger keeps it running.

### 4.6. Job log

- The job writes the tag, the host architecture and the result of each step of
  section 4.5 to its log. It does not write the settings of section 4.1.

### 4.7. Rollback

- To deploy an earlier release, re-run its workflow run in GitHub
  ("Re-run all jobs"). The re-run uses the commit and the tag of that run.
  GitHub allows a re-run for 30 days after the run.
- After 30 days, revert the change in `main` and push a new tag.

## 5. Units

`velourbbs.socket`:

| Section | Line | Reason |
|---|---|---|
| `[Unit]` | `Description=Velour BBS socket` | The name in `systemctl` and in the log. |
| `[Socket]` | `ListenStream=@PORT@` | The port on all addresses, IPv4 and IPv6. |
| `[Install]` | `WantedBy=sockets.target` | `enable` opens the socket when the user manager starts. |

`velourbbs.service`:

| Section | Line | Reason |
|---|---|---|
| `[Unit]` | `Description=Velour BBS` | The name in `systemctl` and in the log. |
| `[Unit]` | `Requires=velourbbs.socket`, `After=velourbbs.socket` | The service gets the socket. |
| `[Service]` | `Type=exec` | `systemctl start` fails when systemd cannot start the program. |
| `[Service]` | `ExecStart=%h/velourbbs/current/velourbbs serve` | The command of the current release. |
| `[Service]` | `WorkingDirectory=%h/velourbbs/current/station` | The BBS directory ([bbs-toml](bbs-toml.md), section 2). |
| `[Service]` | `Restart=on-failure` | systemd starts the host again after a failure. |
| `[Service]` | `MemoryMax=128M`, `TasksMax=64`, `LimitNOFILE=1024` | The limits of [setup-guide](setup-guide.md), section 5. |
| `[Install]` | `WantedBy=default.target` | `enable` starts the host when the user manager starts. |

- The host gets the socket by socket activation ([telnet-term](telnet-term.md), section 10).
  `telnet.listen` of `station/bbs.toml` has no effect.
- A port below 1024 needs `net.ipv4.ip_unprivileged_port_start` on the host.
  The host setup is not in this specification.

## 6. Libraries

- The workflow uses only the actions `actions/checkout` and `actions/setup-go`.
  GitHub maintains them and documents each breaking change in their releases.
  Each action is pinned to a full commit SHA, with the version in a comment.
- The release uses `gh`, and the deploy uses OpenSSH and `tar`, of the runner image.
  Another action for SSH or for releases is not used: it would get the SSH key.

## 7. Target implementations

- The runner is `ubuntu-latest` of GitHub.
- The host has systemd 240 or later with cgroup v2 ([setup-guide](setup-guide.md), section 5.1).
- `MemoryMax` and `TasksMax` of a user unit work only when `user@.service` delegates
  the `memory` and `pids` controllers. Check it on the host:
  `systemctl cat user@.service` must have `memory` and `pids` in `Delegate=`.
  systemd 259 of Ubuntu has `Delegate=pids memory cpu`.
- The user exists on the host, with linger and the public key in `~/.ssh/authorized_keys`.

## 8. Acceptance

- `systemd-analyze verify` gives no error for the two units, with `@PORT@` replaced
  and with `ExecStart` set to the path of a built `velourbbs`.
- `shellcheck` gives no warning for `deploy/install.sh`.
- A test of `install.sh` runs it with a stub `systemctl` in a temporary home directory.
  It checks the layout of section 4.5 and that a failed `velourbbs check` keeps `current`.
- A push of a pre-release tag creates a GitHub pre-release with 4 binaries and `SHA256SUMS`.
  The home BBS then shows the tag in `bbs.version`.
- A tag on a commit outside `main` fails the job with no release.

## 9. Not in this specification

- The host setup: the user, linger, the SSH key, the firewall.
- Deploy to another host, or of another BBS directory than `station/`.
- Signatures of the release assets.
