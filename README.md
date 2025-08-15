# big bang

thy dots and scripts

## Installation Script

I aim to keep my setup fully reproducible, provided I can create a new user on an existing system (MacOS, Debian, or NixOS). The process starts with
`big_bang.sh`, which handles all system-wide configuration and downloads a pinned version of Go.

Once the initial setup is complete, it runs: `go run ./big_bang.go`. This script manages my dotfiles and user-level dependencies—essentially, my core
development tools.

The dotfiles directory is a mirror of the home directory, but syncing is one-way: it creates or overwrites files in `$HOME` without deleting anything that isn’t
in dotfiles. This means that if you remove a file from dotfiles, it will remain in the actual home directory until you delete it manually. This approach avoids
using symlinks altogether.

A notable detail in big_bang.go is a custom 400-line logger I wrote, inspired by Zerolog, offering similar performance with zero heap allocations.

## Dependencies

I'd like to write a comprehensive discussion on my core dev tools, but for now I'll just say that if I could only have two, I'd need `nvim` and `fzf`.

## Theming

### Font - Nerd Font JetBrains Mono
I originally used Iosevka for its thinner typeface, which allowed more columns to fit in my terminal—useful when splitting the window vertically in tmux. After
changing my workflow to use only full-width windows, that benefit became irrelevant. With JetBrains Mono, I can already fit over 200 columns on my screen, so
font choice no longer affects my workflow.

It doesn’t even matter if the font is a Nerd Font, as I’m not particularly fond of code ligatures—though I’m not bothered enough to disable them. If I ever
needed a fallback, I’d pick Hurmit Nerd Font, as its braces, parentheses, and brackets are highly distinguishable from one another.

### Color Palette  - Rose Pine

Yes, even with colorschemes, there’s value in minimizing “dependencies”—both in how syntax highlighting is implemented and in deciding what actually needs
highlighting. I now stick to simple regex-based highlighting. Treesitter is powerful, but it’s a heavy dependency with performance issues on large files. LSPs
share similar drawbacks, and I’ve moved away from them entirely.
