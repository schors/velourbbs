# Idea: deploy-limits

Backlog. It adds network limits outside the host process to the
[setup guide](../specs/setup-guide.md).
It does not change [telnet-term](../specs/telnet-term.md).

## Goal

The deployment limits connections and traffic before they reach the host process.

## Decisions

- The host limits of telnet-term §9.2 and §9.3 stay as they are. Do not extend them.
- nftables limits the connections for each address (`ct count`) and the connect rate.
- `tc` shapes traffic (QoS), if needed.

## Open questions

- Which nftables limits repeat a host limit.
- Behind a PROXY protocol sender, nftables sees the proxy address.
  The per-address rules then belong on the proxy.
- A connection that nftables drops gets no text line (telnet-term §9.2).
- How to check the nftables rules without root.

## Facts

- nftables works only on Linux and needs root.
- systemd `MaxConnectionsPerSource` needs `Accept=yes` socket activation.
  It does not apply to `velourbbs serve`.
