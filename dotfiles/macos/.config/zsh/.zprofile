# Place path exports in .zprofile - https://stackoverflow.com/a/34244862
# Zsh on Arch [and OSX] sources /etc/profile – which overwrites and exports PATH – after having sourced ~/.zshenv
# It should be placed instead in .zprofile to be loaded after the fact.
export PATH="$CARGO_HOME/bin:$PATH"
# Put BIG_BANG_BIN last for it to take priority.
export PATH="$BIG_BANG_BIN:$PATH"

eval "$(/opt/homebrew/bin/brew shellenv)"

fish
