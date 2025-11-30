if status is-interactive
        abbr --add --set-cursor=!! bx bash -c \'!!\'
        abbr --add n nvim .

        # Git
        abbr --add l lazygit
        abbr --add gco git clone --depth=1 --no-single-branch
        abbr --add gfb git clone --filter=blob:none 
        abbr --add --set-cursor=!! gmt git commit --allow-empty --message \"!!\"

        abbr --add stdrs cd $(rustc --print sysroot)/lib/rustlib/src/rust/library/
        abbr --add stdgo cd $(go env GOROOT)

        abbr --add --set-cursor=!! goto 'set dir $HOME/!!/; 
        cd (echo $dir(string join \n (echo ./) (fd --type directory --max-depth 1 --base-directory $dir) | fzf)) 2>/dev/null
        and nvim .'
        abbr --add c 'set dir $HOME/code/; 
        cd (echo $dir(string join \n (echo ./) (fd --type directory --max-depth 1 --base-directory $dir) | fzf)) 2>/dev/null; 
        and nvim .'
        bind ctrl-h backward-kill-word
        # bind ctrl-w : # i dont know how to erase builtin binds
end
