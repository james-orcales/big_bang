# big bang

thy dots and scripts

## Installation Script

I aim to keep my setup fully reproducible, provided I can create a new user on an existing system (MacOS, Debian, or NixOS). The process starts with
`bootstrap.lua`, which handles all system-wide configuration and downloads a pinned version of Go.

Once the initial setup is complete, it runs: `go run ./big_bang.go` This script manages my dotfiles and user-level dependencies—essentially, my core development
tools.

The dotfiles directory is a mirror of the home directory, but syncing is one-way: it creates or overwrites files in $HOME without deleting anything that isn’t
in dotfiles. This means that if you remove a file from dotfiles, it will remain in the actual home directory until you delete it manually. This approach avoids
using symlinks altogether.

A notable detail in `big_bang.go` is a custom 400-line logger I wrote, inspired by Zerolog, offering similar performance with zero heap allocations.

## Dotfiles Management

## Theming

### Font - Nerd Font JetBrains Mono

I originally used Iosevka for its thinner typeface, which allowed more columns to fit in my terminal—useful when splitting the window vertically in tmux. After
changing my workflow to use only full-width windows, that benefit became irrelevant. With JetBrains Mono, I can already fit over 200 columns on my screen, so
font choice no longer affects my workflow.

It doesn’t even matter if the font is a Nerd Font, as I’m not particularly fond of code ligatures—though I’m not bothered enough to disable them. If I ever
needed a fallback, I’d pick Hurmit Nerd Font, as its braces, parentheses, and brackets are highly distinguishable from one another.

### Color Palette  - Rose Pine

I've spent weekends exploring different colorschemes. Now I fall back to rose pine because of that dark green + pink combo--high contrast, colorful, and
minimal. In Neovim however, my setup is simple: comments are gray, TODO comments are red, strings yellow, and everything else stays white.

Even with colorschemes, I think there’s value in minimizing “dependencies”—both in how much syntax you choose to highlight and in the tools you rely on.
Treesitter is powerful, but it’s a heavy dependency and is sluggish on large files. LSPs have similar issues, and over time I’ve moved away from using them
altogether.
