# Idea: deploy-limits

Backlog. It adds deployment files or documentation for `bbs serve`.
It does not change [telnet-term](../specs/telnet-term.md).

## Goal

The deployment limits connections and resources outside the host process.

## Decisions

- The host limits of telnet-term §9.2 and §9.3 stay as they are. Do not extend them.
- A systemd unit runs `bbs serve` with `MemoryMax`, `TasksMax` and `LimitNOFILE`.
- nftables limits the connections for each address (`ct count`) and the connect rate.
- `tc` shapes traffic (QoS), if needed.

## Open questions

- Which deployment limits repeat a host limit.
- Behind a PROXY protocol sender, nftables sees the proxy address.
  The per-address rules then belong on the proxy.
- A connection that nftables drops gets no text line (telnet-term §9.2).
- Example files in the repository, or documentation only.

## Facts

- nftables works only on Linux and needs root.
- systemd `MaxConnectionsPerSource` needs `Accept=yes` socket activation.
  It does not apply to `bbs serve`.
