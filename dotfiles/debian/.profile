# if running bash
if [ -n "$BASH_VERSION" ]; then
    # include .bashrc if it exists
    if [ -f "$HOME/.bashrc" ]; then
	. "$HOME/.bashrc"
    fi
fi

# set PATH so it includes user's private bin if it exists
if [ -d "$HOME/.local/bin" ] ; then
    PATH="$HOME/.local/bin:$PATH"
fi

export PATH="$HOME/.config/sway/bin:$PATH"
export PATH=$PATH:/usr/local/go/bin
export PATH=$PATH:~/go/bin
export PATH=$PATH:~/.cargo/bin
export XDG_CONFIG_HOME=$HOME/.config

export EDITOR=nvim
export XDG_CURRENT_TYPE=wayland
export MOZ_ENABLE_WAYLAND=1
export QT_QPA_PLATFORM=wayland

export ODIN_ROOT="$HOME/build/odin"
source $XDG_CONFIG_HOME/fzf/config

export GPG_TTY=$(tty)
gpgconf --launch gpg-agent


eval "$(keychain --eval --quiet id_ed25519 work_ed25519)"

SOCK="/tmp/ssh-agent-$USER"
if test $SSH_AUTH_SOCK && [ $SSH_AUTH_SOCK != $SOCK ]
then
    rm -f /tmp/ssh-agent-$USER
    ln -sf $SSH_AUTH_SOCK $SOCK
    export SSH_AUTH_SOCK=$SOCK
fi

if [ "$(tty)" = "/dev/tty1" ] ; then
    export XDG_CURRENT_DESKTOP=sway
    exec sway
fi
