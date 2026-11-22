# 2. The BBS directory

A "BBS directory" holds the files of one BBS. This guide uses `/srv/velourbbs`.
Your user owns it. The BBS runs as your user too (section 5):
it reads the files and writes its data file to this directory.
Create the directory and copy the example BBS "Lamp" to it. In the source repository, run:

```
sudo install -d -o "$USER" -g "$(id -gn)" /srv/velourbbs
cp -r examples/lamp/en/. /srv/velourbbs/
```

Edit the files as your user. Do not use `sudo` for the files of the BBS directory.

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
| `screens/log.bbt` | The keeper's log: the list of the notes. |
| `screens/goodbye.bbt` | The screen after the end of the session. |
| `content/lamp.md` | A text in Markdown: a short history of the lamps. |
| `content/rules.txt` | A plain text: the rules of the station. |
| `content/log/` | The notes of the keeper's log, one Markdown file for each note. |

To change the screens, read the [BBT guide](../bbt/README.md).

Run each `velourbbs` command in the BBS directory, as your user.
Full rules: [bbs-toml](../../../specs/bbs-toml.md), section 2.

## Result

```
cd /srv/velourbbs
velourbbs check
```

The command gives no output.
