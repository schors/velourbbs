# Idea: session and BBS fields on screens

Backlog. It extends the field registry of [bbt-format](../specs/bbt-format.md), section 4.1.

## Goal

A screen shows the user the data of the session and of the BBS through BBT fields.

## Decisions

- New fields in the field registry:
  - The IP address of the client. Behind a PROXY protocol sender, the address
    from the PROXY header.
  - The online time of the current session. A string `hh:mm`.
  - The BBS version. The git tag that the build embeds. A build without a tag shows `dev`.
  - The number of open sessions, with the current session. An integer.
  - The current date and time of the server. The theme in `screens/screens.toml`
    gives the format. The format is a layout of the Go package `time`: `2006-01-02 15:04`.
  - The BBS uptime: `3d 04:17`. Less than one day gives `04:17`.
  - The BBS name. A new key in `bbs.toml`.
- The host takes the value when it renders the screen. The next show of the screen
  gives the new value. [live-fields](live-fields.md) is the update on an open screen.
- No node number.

## Open questions

- The values of the session fields in `velourbbs preview`, which has no telnet.
- The default of the BBS name.
- One date and time format in the theme, or two.
- The online time format after 99 hours.
