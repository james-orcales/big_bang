if status is-interactive
        abbr --add --set-cursor=!! bx bash -c \'!!\'
        abbr --add --position anywhere errout '3>&1 1>/dev/null 2>&3 |'

        abbr --add ls ls --color=auto --group-directories-first 
        abbr --add ll ls --color=auto --group-directories-first -l
        abbr --add la ls --color=auto --group-directories-first -a

        abbr --add n nvim .
        abbr --add cfg /usr/bin/git --git-dir=$HOME/.cfg/ --work-tree=$HOME
        abbr --add l lazygit
        abbr --add lc lazygit --git-dir=$HOME/.cfg/ --work-tree=$HOME
        abbr --add gco git clone --depth=1 --no-single-branch

        abbr --add t zoxide_cd

        bind ctrl-space,c "fzf_nvim_oil $XDG_CONFIG_HOME" repaint
        bind ctrl-space,d "fzf_nvim_oil $HOME/Documents" repaint
        bind ctrl-space,p "fzf_nvim_oil $HOME/personal/git" repaint
        bind ctrl-space,s "fzf_nvim_oil $HOME/personal/git/scratch" repaint
        bind ctrl-space,w "fzf_nvim_oil $HOME/work/git" repaint
        bind ctrl-h backward-kill-word
        # bind ctrl-w : # i dont know how to erase builtin binds
end
