if status is-interactive
        abbr --add --set-cursor=!! bx bash -c \'!!\'
        abbr --add --position anywhere errout '3>&1 1>/dev/null 2>&3 |'

        abbr --add n nvim .
        abbr --add cfg /usr/bin/git --git-dir=$HOME/.cfg/ --work-tree=$HOME
        abbr --add l  lazygit
        abbr --add lc lazygit --git-dir=$HOME/.cfg/ --work-tree=$HOME
        abbr --add gco git clone --depth=1 --no-single-branch
        abbr --add --set-cursor=!! gmt git commit --allow-empty --message \"!!\"

        abbr --add t zoxide_cd

        bind ctrl-space,c "fzf_cd $HOME/.config              && nvim ."             repaint
        bind ctrl-space,d "fzf_cd $HOME/Documents            && nvim ."             repaint
        bind ctrl-space,p "fzf_cd $HOME/personal/git         && tmux_preset_layout" repaint
        bind ctrl-space,s "fzf_cd $HOME/personal/git/scratch && tmux_preset_layout" repaint
        bind ctrl-space,w "fzf_cd $HOME/work/git/            && tmux_preset_layout" repaint
        bind ctrl-h backward-kill-word
        # bind ctrl-w : # i dont know how to erase builtin binds
end
