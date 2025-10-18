if vim.fn.fnamemodify(vim.fn.getcwd(), ":~") == "<REPO>" then
        vim.opt.makeprg = "vendor/Odin/odin check"
end
vim.bo.errorformat = ",%f(%l:%v) %m"
