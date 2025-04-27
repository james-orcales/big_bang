function nvim_oil -d "opens the passed dir in neovim oil"
        if not test -d $argv
                echo "nvim_oil: arg is not a dir. got: $argv"
                return 1
        end
        cd (realpath $argv)
        and nvim +Oil
end
