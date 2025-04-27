if status is-interactive
        abbr --add --position command ls ls --color=auto --group-directories-first 
        abbr --add --position command ll ls --color=auto --group-directories-first -l
        abbr --add --position command la ls --color=auto --group-directories-first -a

        abbr --add --position command n nvim .
        abbr --add --position command cfg /usr/bin/git --git-dir=$HOME/.cfg/ --work-tree=$HOME
        abbr --add --position command lg lazygit
        abbr --add --position command lgc lazygit --git-dir=$HOME/.cfg/ --work-tree=$HOME

        abbr --add --position command t zoxide_cd

        bind ctrl-space,c "fzf_nvim_oil $XDG_CONFIG_HOME" repaint
        bind ctrl-space,d "fzf_nvim_oil $XDG_CONFIG_HOME/Documents" repaint
        bind ctrl-space,p "fzf_nvim_oil $HOME/personal/git" repaint
        bind ctrl-space,s "fzf_nvim_oil $HOME/personal/git/scratch" repaint
        bind ctrl-space,w "fzf_nvim_oil $HOME/personal/work/git" repaint
        bind ctrl-h backward-kill-word
        bind ctrl-w : # i dont know how to erase builtin binds
end
