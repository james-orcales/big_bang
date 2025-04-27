function fzf_nvim_oil
        if not test -d $argv
                echo "fzf_nvim_oil arg is not a dir got:" $argv 
                return 1
        end

        set -f DIR (begin; echo "/"; fd -L -d 1 -t d . $argv; end |
                xargs -n 1 basename |
                fzf
        )
        and nvim_oil (echo -s -- $argv "/" $DIR)
end
