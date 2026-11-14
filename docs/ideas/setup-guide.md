# Idea: setup-guide

Backlog. It adds a guide to the setup of a BBS.
It comes after [bbs-toml-spec](bbs-toml-spec.md).
It does not change [telnet-term](../specs/telnet-term.md).

## Goal

The sysop sets up, checks and runs a BBS without the specifications.
The deployment limits connections and resources outside the host process.

## Decisions

- The reader is the sysop.
- The guide has: the BBS directory and `bbs.toml`, the commands (`velourbbs check`,
  `velourbbs preview`, `velourbbs preview --keys`, `velourbbs art import`,
  `velourbbs serve`) and the deployment.
- The guide refers to the `bbs.toml` specification for the full rules.
- The guide is in English and Russian. English is the source, Russian is a translation.
  Change both in the same commit.
- An example BBS will exist as a separate issue. The guide takes it into account.
- The host limits of telnet-term §9.2 and §9.3 stay as they are. Do not extend them.
- A systemd unit runs `velourbbs serve` with `MemoryMax`, `TasksMax` and `LimitNOFILE`.
- nftables limits the connections for each address (`ct count`) and the connect rate.
- `tc` shapes traffic (QoS), if needed.

## Open questions

- The place of the guide files.
- Which deployment limits repeat a host limit.
- Behind a PROXY protocol sender, nftables sees the proxy address.
  The per-address rules then belong on the proxy.
- A connection that nftables drops gets no text line (telnet-term §9.2).
- Example files in the repository, or documentation only.

## Facts

- nftables works only on Linux and needs root.
- systemd `MaxConnectionsPerSource` needs `Accept=yes` socket activation.
  It does not apply to `velourbbs serve`.
