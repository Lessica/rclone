# Rclone Minimal

This is my personal minimal [Rclone](https://rclone.org/) repo. It is aimed at being as minimal as possible, removing all backends and commands that I do not use.

## Major Changes

- Only following backends are preserved:
  - Google Drive
  - Dropbox
  - Local
  - One Drive
  - WebDAV
  * … **Other backends including virtual ones are removed.**
- All commands related to virtual backends are removed.
- All mount commands are removed.
- All documents, auto-completions and test commands are removed.
- Remote control and WebGUI are both removed.
- `libhttp` and `rclone serve` are both removed.
- Self-update is disabled.

## Why?

I want to use `rclone` on my mobile phone and I do not need these strange storage providers at all.

These features are not good for my little phone, too: Mount as FS / Remote Control / Terminal Auto-Completion / Test Commands

Remote control? WebGUI? `rclone serve`? Oh my god why you guys added so many features to a command line tool?
