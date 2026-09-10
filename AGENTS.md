# Before

Before reading this file read ~/ai/AGENTS.md

# Project rules

- This is a CLI tool
- Is written entirely in golang
- It must be a full TUI (not a command-oriented CLI, not a hybrid)

# UX contract

Layout:

```
+------------------------+--------------------+
| [01:30]  Item1         |                    |
| [00:00]  Item2         |    00:00:00        |  large digits, blink if zero
| [00:45]  Item3         |      [00:29]       |  overtime, normal type
|                        +--------------------+
|                        | notes              |
+------------------------+--------------------+
```

The left pane is the full-height item list.
Each row is `[hh:mm] title`, remaining time for that item.
White: not done, not running.
Green: running and remaining greater than zero.
Red: running and remaining is zero (overtime).
Gray with strikethrough title: done.
The `[hh:mm]` duration stays unstruck.
Running list durations keep counting down while those items are ticking.
The selected row is highlighted even when it is not a runner.

The top-right pane is the big clock for the **selected** item.
Empty list stays `--:--:--`.
It must be block digits showing `hh:mm:ss`, not a status-bar timestamp.
When remaining hits `00:00:00`, the big clock stays at `00:00:00` and blinks while that selected item is running.
The terminal bell must ring twice on that transition for each runner that hits zero (stderr BEL, not every overtime tick).
Overtime then counts up from zero under the digits as `[hh:mm]` in normal size.
Below that, `Invested time: hh:mm` shows how long that clock item has been running (second resolution, display hours and minutes only).
Invested time keeps increasing after remaining hits zero.
Pause freezes that item's invested counter.

The bottom-right pane is notes for the **selected** item.
Notes are raw text while the notes pane is focused.
Notes render WhatsApp-style `*bold*`, `_italic_`, and `~strike~` when the notes pane is not focused.
Enter in notes inserts a newline.

Running (Space, ticking) and selected (highlight, notes, big clock) are independent.
Several items can run at the same time.

## Navigate (list focused)

- Up/down: change selection
- Tab / Shift-Tab: list <-> notes
- Enter: start edit on the selected row
- Space: start or pause the selected item (list only, navigate only)
- Space on a white item starts it and leaves other runners ticking
- Space on a green or red item pauses that item only
- Space on a done item is ignored
- No cap on how many items can run
- `n`: insert a new item at the **top** of the list (`[00:00]`, empty title) and enter edit on the duration field
- `-`: move the selected item down
- `=` / `+`: move the selected item up
- `x`: toggle done (strikethrough title). Marking a running item done stops that clock only. Undoing done does not auto-resume.
- `d`: open a confirm modal: "do you want to delete this item?" with Yes / No / Done (default Yes)
- Delete modal: Left/Right or Tab cycle choices, Enter accepts, Esc is No
- Yes deletes the item (other runners keep ticking). No cancels. Done marks the item complete without deleting (stops that runner only).
- `q` / ctrl+c: quit (flush JSON)

## Edit (list)

- Tab / Shift-Tab: duration <-> title (not notes)
- Duration is a 4-digit mask `hh:mm` (hours 00-99)
- Colon is fixed; left/right skip it
- Digit keys overwrite the current slot and advance (skip colon)
- Minutes tens digit only accepts 0-5
- Backspace sets that digit to 0 and moves left (does not delete the colon)
- Left/right: cursor in the current field
- Up/down: other item, stay in edit (same field type)
- Esc: back to navigate
- Space: insert a space in the title; ignored in `[hh:mm]`

## Persistence

State lives at `~/.timebox_me/state.json`.
Persist items, remaining, overtime, invested time, notes, done flag, running flag per item, and selected item.
Missing file means an empty list.
Legacy files with top-level `active_id` and `running` load that one item as running when `running` is true.
