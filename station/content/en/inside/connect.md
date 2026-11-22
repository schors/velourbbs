# Connection

## What it does

Velour accepts a TCP connection, finds the address of the guest and checks the limits.
Only then does telnet start.

## How

### Port

On Linux, systemd holds the port, not Velour: this is socket activation.
systemd opens the port and starts Velour with the variables `LISTEN_PID` and `LISTEN_FDS=1`.
The open port comes as file descriptor 3, by the rules of `sd_listen_fds`
([systemd documentation](https://www.freedesktop.org/software/systemd/man/latest/sd_listen_fds.html)).
Without systemd, Velour listens on the address `telnet.listen` of `bbs.toml` itself.

### Address behind a proxy

When a proxy stands before Velour, the address of the guest comes in a PROXY protocol header
([HAProxy specification](https://www.haproxy.org/download/3.1/doc/proxy-protocol.txt)).
Velour reads the header only from the addresses of the list `proxy` in `bbs.toml`.

- Version 1 is a text line such as `PROXY TCP4 <address> <address> <port> <port>`,
  at most 107 bytes.
- Version 2 is a binary header: 16 bytes and up to 536 bytes of addresses and TLVs.
  Velour skips the TLVs.

HAProxy, nginx, Traefik and stunnel send the header.

### Connection limits

- `max_conns`: the connections of the whole server.
- `max_conns_per_addr`: the open connections from one address.
- `max_connects_per_min`: the connection attempts from one address in 60 seconds.
  Velour counts them in steps of 1 second and keeps up to 10000 addresses.

One address is one IPv4 address or one IPv6 /64 network.
An IPv4 address inside IPv6 (`::ffff:a.b.c.d`) counts as the IPv4 address.
The addresses of the list `allow` have no address limits.

## What it gets

- The address of the guest. The limits, the log and the field `session.ip`
  on the screen "About the station" see it.
- The headers `UNKNOWN` (version 1), `LOCAL`, `UNSPEC` and `UNIX` (version 2) give no address.
  Then Velour takes the address of the proxy, and the address limits do not apply.
- A connection over a limit gets one line `Too many connections. Try again later.`
  and closes.

Before the close, Velour closes its write side and reads the input of the guest for up to 1 second.
A close with unread input makes TCP send a reset, and the client can lose the last text.

## Limits

- The PROXY header and the terminal probe must end within `probe_timeout` from the connection.
- An invalid or missing header closes the connection with no reply.
- The proxy must send the header at once. Velour writes first, so a proxy
  that waits for client data does not work. AWS NLB does this by default.
- Now the only connection is telnet without encryption. SSH and telnet over TLS are planned.
