vim.diagnostic.config({ virtual_lines = { current_line = true } })


vim.opt.laststatus = 3
vim.api.nvim_create_autocmd({ "BufWinEnter" }, {
	desc = "Get git branch of opened buffer for statusline",
	group = vim.api.nvim_create_augroup("statusline-git-branch", { clear = true }),
	callback = function()
                local file_path = vim.fn.expand("%:p:h")
                local prefix = "oil://"
                if string.sub(file_path, 1, #prefix) == prefix then
                        file_path = string.sub(file_path, #prefix + 1)
                end
                local branch = vim.trim(vim.fn.system("git -C " .. file_path .. " branch --show-current 2>/dev/null"))
                vim.b.git_branch = #branch > 0 and string.format("git:(%s)", branch) or ""
	end,
})
vim.opt.statusline = 
        "%<%{expand('%:~')}(%(%l:%v%))" .. 
        " %{exists('b:git_branch') ? b:git_branch : ''}" ..
        " %h%w%{&modified ?  '[MODIFIED]' : ''}%r" .. 
        " %=Rune=%B" .. 
        " Byte_Index=%o" ..
        " %q" ..
        " %P"


vim.opt.smartcase  = true
vim.opt.ignorecase = true


vim.opt.number         = true
vim.opt.relativenumber = true


vim.opt.tabstop     = 8
vim.opt.softtabstop = 8
vim.opt.shiftwidth  = 8
vim.opt.expandtab   = true


vim.opt.smartindent = true


vim.opt.wrap = false


vim.opt.swapfile = false
vim.opt.backup   = false
vim.opt.undodir  = os.getenv("HOME") .. "/.local/share/nvim/undodir"
vim.opt.undofile = true


vim.opt.hlsearch  = false
vim.opt.incsearch = true


vim.opt.termguicolors = true


vim.opt.scrolloff = 1


vim.opt.signcolumn = "yes"


vim.opt.isfname:append("@-@")


vim.opt.updatetime = 750


vim.opt.colorcolumn = "160"
vim.opt.textwidth   = 160
vim.opt.wrapmargin  = 1
vim.opt.formatoptions:append("t")


vim.opt.completeopt = { "menu", "menuone", "noselect" }
vim.opt.fillchars   = { eob = " " }
