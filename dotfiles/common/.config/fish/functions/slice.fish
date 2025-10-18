function slice -a bytes lower operator upper 
        switch $operator 
        case '..='
        case '*'
                echo invalid operator
                exit 1
        end
        head -c $upper $bytes | tail -c (math $upper - $lower)
end
