# Task 3. The header

Each screen of Lamp starts with the same 5 rows: the "header".
Write them once in a fragment, and include the fragment in each screen.

- A fragment is `screens/_<name>.bbt`. It has no front matter.
- The line `{>_<name>}` includes the fragment: the BBS puts the lines of the fragment in its place.
  This line must have no other text.
- After the include, the style is the style before the include.

Full rules: [bbt-format](../../../specs/bbt-format.md), sections 3.3 and 3.4.

In `screens/main.bbt`, replace rows 1–5 with one line:

```
+++
[[item]]
key = "G"
do = "logoff"
+++
{>_head}
       Good evening. It's quiet and warm here.
```

Move the 5 rows to the fragment:

`examples/lamp/en/screens/_head.bbt`:
```

  {!line}────────────────────────────────────────────────────────────────────────────{!}

       {!title}Lamp{!}  {!orn}☼ {!dim}a BBS by lamplight{!}

```

## Result

- `velourbbs check` gives no output.
- `velourbbs preview main` gives the same output as in task 2. To compare, run:

```
velourbbs preview main > task3.txt
diff task2.txt task3.txt
```

`diff` gives no output. On Windows, use `fc task2.txt task3.txt`.
