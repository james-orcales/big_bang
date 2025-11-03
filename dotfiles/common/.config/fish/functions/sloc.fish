# Clone a git repository then count its lines of code
function sloc -a git_url
        if not command -v tokei
                echo "error: tokei is not installed"
                return 1
        end
        if test $git_url = ""
                echo "error: missing git url"
                return 1
        end
        set name (basename $git_url .git)
        git clone --depth=1 --single-branch $git_url
        and tokei $name
end

