# 2. The BBS directory

A "BBS directory" holds the files of one BBS. This guide uses `/srv/velourbbs`.
Copy the example BBS "Lamp" to it. In the source repository, run:

```
sudo cp -r examples/lamp/en /srv/velourbbs
sudo chmod -R a+rX /srv/velourbbs
```

`/srv/velourbbs` must not exist before the copy.
`chmod` makes the files readable and the directories readable and searchable for each user.
The BBS runs as the user `velourbbs` (below). This user must read the files.
The files belong to root. Edit them as root, for example with `sudoedit`.

The files of Lamp:

| File | Content |
|---|---|
| `bbs.toml` | The settings of the host (section 3). In Lamp, it is empty: the defaults. |
| `screens/screens.toml` | The settings of the screens: the width, the styles and other values. |
| `screens/_head.bbt` | The header. Each screen includes it. |
| `screens/main.bbt` | The first screen: the menu. |
| `screens/about.bbt` | The text about the BBS. |
| `screens/term.bbt` | The values of the terminal. |
| `screens/texts.bbt` | The list of the texts of the reading room. |
| `screens/reader.bbt` | The screen that shows a text of the list. |
| `screens/goodbye.bbt` | The screen after the end of the session. |
| `content/lamp.md` | A text in Markdown: a short history of the lamps. |
| `content/rules.txt` | A plain text: the rules of the station. |

To change the screens, read the [BBT guide](../bbt/README.md).

## The BBS user

The BBS runs as the system user `velourbbs`, without privileges.
The user has no login and no home directory.
The BBS writes its data file to the directory `/var/lib/velourbbs` (section 3).
Create the user and the directory:

```
sudo useradd --system --no-create-home --shell /usr/sbin/nologin velourbbs
sudo install -d -o velourbbs -g velourbbs /var/lib/velourbbs
```

Run each `velourbbs` command in the BBS directory.
Run `velourbbs check` and `velourbbs serve` as the user `velourbbs`:
`sudo -u velourbbs velourbbs check`. Root can read a file
that the user `velourbbs` cannot read.
Full rules: [bbs-toml](../../../specs/bbs-toml.md), section 2.

## Result

```
cd /srv/velourbbs
sudo -u velourbbs velourbbs check
```

The command gives no output.
