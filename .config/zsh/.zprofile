# set PATH so it includes user's private bin if it exists
if [ -d "$HOME/bin" ] ; then
    PATH="$HOME/bin:$PATH"
fi

# set PATH so it includes user's private bin if it exists
if [ -d "$HOME/.local/bin" ] ; then
    PATH="$HOME/.local/bin:$PATH"
fi

export PATH="$HOME/.config/sway/bin:$PATH"
export PATH=$PATH:/usr/local/go/bin
export PATH=$PATH:~/go/bin
export XDG_CONFIG_HOME=$HOME/.config

export EDITOR=nvim
export XDG_CURRENT_TYPE=wayland
export MOZ_ENABLE_WAYLAND=1
export QT_QPA_PLATFORM=wayland

export ODIN_ROOT="/home/amesaine/build/odin"
source $XDG_CONFIG_HOME/fzf/config

export GPG_TTY=$(tty)
gpgconf --launch gpg-agent

ZSH_AUTOSUGGEST_HIGHLIGHT_STYLE="fg=#999999,bg=#282A36"

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
