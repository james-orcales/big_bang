-- === Colorscheme ===
local palette = {
        ["yellow"] = "#F6C177",
        ["red"] = "#EB6F92",
        ["blue"] = "#9CCFD8",
        ["text_dark"] = "#777777",
}


-- stylua: ignore start
vim.cmd.colorscheme("quiet")
vim.api.nvim_set_hl(0, "Comment",     { fg = palette["text_dark"] })
vim.api.nvim_set_hl(0, "String",      { fg = palette["yellow"]    })
vim.api.nvim_set_hl(0, "Directory",   { fg = palette["blue"]      })
vim.api.nvim_set_hl(0, "Visual",      { bg = "#333333",           })
vim.api.nvim_set_hl(0, "NormalFloat", { bg = "#0A0A0A"            })
vim.api.nvim_set_hl(0, "StatusLine",  { bg = "#111111"            })
vim.api.nvim_set_hl(0, "StatusLine",  { bg = "#111111"            })
vim.api.nvim_set_hl(0, "TODO",        { fg = palette["red"]       })
vim.api.nvim_set_hl(0, "YankSystemClipboard", { bg = "#0000FF", fg = "#000000" })
vim.diagnostic.config({ virtual_lines = { current_line = true } })
-- stylua: ignore end


-- === Options ===
vim.opt.laststatus = 3
vim.opt.statusline = "%<%{expand('%:~')}(%(%l:%v%))"
        .. " %{exists('b:git_branch') ? b:git_branch : ''}"
        .. " %h%w%{&modified ?  '[MODIFIED]' : ''}%r"
        .. " %=Rune=%B"
        .. " Byte_Index=%o"
        .. " %q"
        .. " %P"


vim.opt.smartcase = true
vim.opt.ignorecase = true


vim.opt.number = true
vim.opt.relativenumber = true


vim.opt.tabstop = 8
vim.opt.softtabstop = 8
vim.opt.shiftwidth = 8
vim.opt.expandtab = true


vim.opt.smartindent = true


vim.opt.wrap = false


vim.opt.swapfile = false
vim.opt.backup = false
vim.opt.undodir = os.getenv("HOME") .. "/.local/share/nvim/undodir"
vim.opt.undofile = true


vim.opt.hlsearch = false
vim.opt.incsearch = true


vim.opt.termguicolors = true


vim.opt.scrolloff = 1


vim.opt.signcolumn = "yes"


vim.opt.isfname:append("@-@")


vim.opt.updatetime = 750


vim.opt.colorcolumn = "160"
vim.opt.textwidth = 160
vim.opt.wrapmargin = 1
vim.opt.formatoptions:append("t")


vim.opt.completeopt = { "menu", "menuone", "noselect" }
vim.opt.fillchars = { eob = " " }


-- === Keymap ===
vim.g.mapleader = " "


local os = vim.uv.os_uname().sysname
local xplat_set = vim.keymap.set
if os == "Darwin" then
        xplat_set = function(modes, lhs, rhs, opts)
                lhs = lhs:gsub("[cC]%-", "M-") -- replace control with Option
                vim.keymap.set(modes, lhs, rhs, opts)
        end
end


-- === Movement ===
-- stylua: ignore start
xplat_set({ "n", "v", "o"      }, "<Home>",     "^zH",               { desc = "Jump to first char of current line and screen hug left" })
xplat_set({ "i"                }, "<Home>",     "<ESC>^zHi",         { desc = "Jump to first char of current line and screen hug left" })
xplat_set({ "n", "v", "o"      }, "<C-Home>",   "gg",                { desc = "Jump first line"                                        })
xplat_set({ "n", "v", "o"      }, "<C-End>",    "G",                 { desc = "Jump last line"                                         })
xplat_set({ "n", "v", "o", "i" }, "<M-b>",      "b",                 { desc = "Jump previous word - MacOS"                             })
xplat_set({ "n", "v", "o", "i" }, "<C-Left>",   "b",                 { desc = "Jump previous word"                                     })
xplat_set({ "n", "v", "o", "i" }, "<M-f>",      "w",                 { desc = "Jump next word - MacOS"                                 })
xplat_set({ "n", "v", "o", "i" }, "<C-Right>",  "w",                 { desc = "Jump next word"                                         })
xplat_set({ "n", "v", "o", "i" }, "<S-Left>",   "B",                 { desc = "Jump previous whitespace"                               })
xplat_set({ "n", "v", "o", "i" }, "<S-Right>",  "W",                 { desc = "Jump next whitespace"                                   })
xplat_set({ "n", "v", "i"      }, "<S-Down>",   "<C-D>zz",           { desc = "Page down, center cursor"                               })
xplat_set({ "n", "v", "i"      }, "<C-S-Down>", "<C-D><C-D><C-D>zz", { desc = "Page down multiple, center cursor"                      })
xplat_set({ "n", "v", "i"      }, "<S-Up>",     "<C-U>zz",           { desc = "Page up, center cursor"                                 })
xplat_set({ "n", "v", "i"      }, "<C-S-Up>",   "<C-U><C-U><C-U>zz", { desc = "Page up multiple, center cursor"                        })
xplat_set({ "i", "c"           }, "<C-H>",      "<C-W>",             { desc = "Kill word before cursor"                                })
xplat_set({ "i", "c"           }, "<C-BS>",     "<C-W>",             { desc = "Kill word before cursor"                                })
-- stylua: ignore end


-- Open File Explorer
xplat_set("n", "-", vim.cmd.Ex)


-- Move selection in visual mode
xplat_set("v", "<C-Down>", ":m '>+1<CR>gv=gv")
xplat_set("v", "<C-Up>", ":m '<-2<CR>gv=gv")
vim.keymap.set({ "n", "i" }, "<C-E>", "<ESC>:w<CR>", { desc = "Save File" })


-- Center screen on search result
xplat_set("n", "n", "nzzzv")
xplat_set("n", "N", "Nzzzv")


-- Quickfix
xplat_set("n", "{", "<CMD>:cprevious<CR>")
xplat_set("n", "}", "<CMD>:cnext<CR>")


xplat_set("n", "H", "<nop>")


-- Saving my right pinky
xplat_set("n", "w", "o", { noremap = true, desc = "Remap o to w" })
xplat_set("n", "W", "O", { noremap = true, desc = "Remap O to W" })
xplat_set("n", "o", "", { noremap = true, desc = "Remap o to w" })
xplat_set("n", "O", "", { noremap = true, desc = "Remap O to W" })


-- Escape in insert mode
-- <C-c> is remapped to <Esc> so that exiting insert mode behaves consistently.
-- This is especially useful for visual block insert mode (Ctrl+v + Shift+i),
-- where <Esc> is required to apply the changes to all selected lines.
-- Without this mapping, using <C-c> instead of <Esc> will cancel the block operation.
vim.keymap.set("i", "<C-c>", "<Esc>")


-- Delete without affecting default register
xplat_set("v", "D", [["_d]])


-- Global case-INsensitive search and replace, matching the word under cursor, without confirmation
xplat_set("n", "sr", [[:%s/\<<C-r><C-w>\>/<C-r><C-w>/gI<Left><Left><Left>]])
-- Visual mode case-sensitive search and replace
xplat_set("v", "sr", [[:s///gc<Left><Left><Left><Left>]])


-- Extend gs to 100ms. Useful in macros involving lsp go-to-definition which has a little delay.
xplat_set("n", "gs", "<cmd>sleep 100m<CR>")


vim.keymap.set({ "n", "v" }, "<C-Y>", [["+y]], { desc = "Yank to system clipboard" })
vim.keymap.set("n", "<C-Y><C-Y>", [["+yy]], { desc = "Yank line to system clipboard" })


-- Add character to end of line
xplat_set("n", ",", "mzA,<ESC>`z")


-- === Autocmd ===
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


vim.api.nvim_create_autocmd("TextYankPost", {
        desc = "Highlight when yanking text",
        group = vim.api.nvim_create_augroup("highlight_yank", { clear = true }),
        callback = function()
                local event = vim.v.event
                local yank_to_clipboard = event.regname == "+"
                if yank_to_clipboard then
                        vim.highlight.on_yank({ higroup = "YankSystemClipboard" })
                else
                        vim.highlight.on_yank()
                end
        end,
})


vim.api.nvim_create_autocmd({ "BufWritePre" }, {
        desc = "Format on save",
        group = vim.api.nvim_create_augroup("format_on_save", { clear = true }),
        pattern = { "*.odin", "*.go", "*.py", "*.sh", "*.lua" },
        callback = function(ev)
                local lines = vim.api.nvim_buf_get_lines(ev.buf, 0, -1, false)
                do
                        local ft = vim.bo[ev.buf].filetype
                        local cmd
                        if ft == "lua" then
                                cmd = { "stylua", "-" }
                        elseif ft == "go" then
                                cmd = { "gofumpt" }
                        end
                        if cmd then
                                local input = table.concat(lines, "\n")
                                local result = vim.system(cmd, { stdin = input, text = true }):wait()
                                if result.code == 0 then
                                        lines = vim.split(result.stdout, "\n", { plain = true })
                                        if lines[#lines] == "" then
                                                table.remove(lines, #lines)
                                        end
                                end
                        end
                end
                assert(#lines > 0)


                -- NOTE: I tried Ex commands with regex at first. The problem was that when undoing, the cursor would
                -- jump to the top of the file.
                --
                -- local pos = vim.api.nvim_win_get_cursor(0)
                -- -- https://vim.fandom.com/wiki/Regex_lookahead_and_lookbehind
                -- vim.cmd([[:%s/^\s*\n\{1,}/\r\r]])
                -- vim.api.nvim_win_set_cursor(0, pos)
                local new_lines = {}
                local blank_streak = 0
                for _, line in ipairs(lines) do
                        if line:match("^%s*$") then
                                blank_streak = blank_streak + 1
                        else
                                if blank_streak > 0 then
                                        table.insert(new_lines, "")
                                        table.insert(new_lines, "")
                                        blank_streak = 0
                                end
                                table.insert(new_lines, line)
                        end
                end
                vim.api.nvim_buf_set_lines(ev.buf, 0, -1, false, new_lines)
        end,
})


-- === Dependencies ===
require("mini/splitjoin").setup()


-- vim-easy-align
vim.g.easy_align_ignore_groups = {}
vim.keymap.set("x", " ", "<Plug>(EasyAlign)")


-- undotree
vim.g.undotree_WindowLayout = 4
vim.g.undotree_shortIndicators = 1
vim.g.undotree_SetFocusWhenToggle = 1
vim.keymap.set("n", "<leader>ut", vim.cmd.UndotreeToggle)


do
        local oil = require("oil")
        oil.setup({
                -- default_file_explorer = true,
                columns = { "icon" },
                buf_options = { buflisted = false, bufhidden = "hide" },
                win_options = {
                        wrap = false,
                        spell = false,
                        list = false,
                        foldcolumn = "0",
                },
                delete_to_trash = false,
                prompt_save_on_select_new_entry = true,
                constrain_cursor = "name",
                keymaps = {
                        ["?"] = "actions.show_help",
                        ["<CR>"] = "actions.select",
                        ["<C-C>"] = oil.discard_all_changes,
                        ["-"] = "actions.parent", -- dash
                        ["_"] = "actions.open_cwd", -- underscore
                        ["cd"] = "actions.cd",
                        ["<C-Home>"] = "gg",
                        ["<C-End>"] = "G",
                        ["="] = function()
                                if vim.g.oil_size_column == 1 then
                                        oil.set_columns({ "icon" })
                                        vim.g.oil_size_column = 0
                                else
                                        oil.set_columns({ "icon", "size" })
                                        vim.g.oil_size_column = 1
                                end
                        end,
                        ["Y"] = "actions.yank_entry",
                },
                natural_order = false,
                use_default_keymaps = false,
                view_options = { show_hidden = true },
        })
        vim.keymap.set({ "n" }, "-", oil.open)
end


require("nvim-surround").setup({
        surrounds = {
                ["("] = false,
                ["["] = false,
                ["<"] = false,
        },
        aliases = {
                ["("] = ")",
                ["["] = "]",
                ["<"] = ">",
        },
        keymaps = {
                normal = "s",
                normal_cur = "ss",
                normal_cur_line = "S",
                visual = "s",
                visual_line = "S",
                delete = "ds",
                change = "cs",
                insert = false,
                insert_line = false,
                normal_line = false,
                change_line = false,
        },
})


do
        local leap = require("leap")
        leap.opts.safe_labels = {}
        leap.opts.labels = "setnriaofuplwyqjbmghdzxc"
        leap.opts.max_phase_one_targets = 0
        leap.opts.special_keys.next_group = "<space>"
        vim.keymap.set({ "n", "x", "o" }, "t", "<Plug>(leap)")
        vim.api.nvim_set_hl(0, "LeapBackdrop", { link = "Comment" })
        -- vim.api.nvim_set_hl(0, "LeapLabelDimmed", { link = "" })
        require("leap").init_hl(true)
end


do
        local fzf = require("fzf-lua")
        fzf.setup({
                defaults = {
                        header = false,
                },
                winopts = { backdrop = 100, fullscreen = true },
                lsp = { symbols = { symbol_style = 3 } },
                hls = { normal = "NormalFloat", border = "FloatBorder" },
                keymap = {
                        fzf = {
                                ["ctrl-h"] = "backward-kill-word",
                                ["shift-down"] = "half-page-down",
                                ["shift-up"] = "half-page-up",
                                ["home"] = "first",
                                ["end"] = "last",
                        },
                },
                actions = {
                        files = {
                                true,
                                ["alt-i"] = fzf.actions.toggle_ignore,
                                ["enter"] = nil,
                                ["ctrl-s"] = nil,
                                ["ctrl-v"] = nil,
                                ["ctrl-t"] = nil,
                                ["alt-q"] = nil,
                                ["alt-Q"] = nil,
                                ["alt-h"] = nil,
                                ["alt-f"] = nil,
                        },
                },
        })
        local module_api_search = function()
                programming_language = nil
                local handle = vim.uv.fs_scandir(vim.uv.cwd())
                if handle then
                        while true do
                                local name, t = vim.uv.fs_scandir_next(handle)
                                if not name then
                                        break
                                end
                                if t == "file" then
                                        if name:match("%.go$") then
                                                programming_language = "Golang"
                                                break
                                        elseif name:match("%.odin$") then
                                                programming_language = "Odin"
                                                break
                                        elseif name:match("%.lua$") then
                                                programming_language = "Lua"
                                                break
                                        elseif name:match("%.py$") then
                                                programming_language = "Python"
                                                break
                                        end
                                end
                        end
                end
                if programming_language == nil then
                        fzf.live_grep()
                        return
                end
                local items = { "Function", "Type", "Variables", "_Function", "_Type", "Any" }
                fzf.fzf_exec(items, {
                        prompt = string.format("Search Package (%s) > ", programming_language),
                        actions = {
                                ["default"] = function(selected, opts)
                                        if selected == nil then
                                                return
                                        end
                                        selected = selected[1]
                                        if selected == "Any" then
                                                fzf.live_grep()
                                                return
                                        end
                                        local pattern, ripgrep_options
                                        if programming_language == "Golang" then
                                                if selected == "Function" then
                                                        local func = [[^func +]]
                                                        local receiver = [[(?:\(\w+ +\*?\w+\))? *]] -- optional
                                                        local identifier = [[[A-Z]\w+]]
                                                        local generics = [[(?:\[.*?\])?]]
                                                        local signature = [[\(.*?\) +]]
                                                        pattern = func .. receiver .. identifier .. generics .. signature
                                                elseif selected == "Type" then
                                                        pattern = [[^type +[A-Z]\w* +]]
                                                elseif selected == "Variables" then
                                                        ripgrep_options = "--multiline"
                                                        -- only matches file scope
                                                        local single_line = [[^(?:var|const) +[A-Z]\w+]]
                                                        local multiline = [[(?s)^(?:var|const) \(.*?^\)]]
                                                        pattern = string.format("(?:%s|%s)", single_line, multiline)
                                                elseif selected == "_Function" then
                                                        pattern = [[^func +.*]]
                                                elseif selected == "_Type" then
                                                        pattern = [[^type +\w+ +]]
                                                end
                                                pattern = pattern .. " -- !*test*"
                                        elseif programming_language == "Odin" then
                                                if selected == "Function" then
                                                        pattern = [[^\w+ +:: +proc]]
                                                elseif selected == "Type" then
                                                        pattern = [[^\w+ +:: +(?:struct|union|enum|distinct)]]
                                                end
                                                pattern = pattern .. " -- !*test*"
                                        elseif programming_language == "Lua" then
                                                if selected == "Function" then
                                                        pattern = [[\w+ += +function\(|function +\w+\(|\w+ += +def\(]]
                                                end
                                        end
                                        fzf.live_grep({
                                                search = pattern,
                                                rg_opts = ripgrep_options,
                                                no_esc = true,
                                                -- Error: unable to init vim.regex
                                                -- https://github.com/ibhagwan/fzf-lua/issues/1858#issuecomment-2689899556
                                                -- The message is mostly informational, this happens due to the
                                                -- previewer trying to convert the regex to vim magic pattern (in
                                                -- order to highlight it), but not all cases can be covered so the
                                                -- previewer will highlight the cursor column only (instead of the
                                                -- entire pattern).
                                                silent = true,
                                        })
                                end,
                        },
                })
        end


        vim.keymap.set("n", "<C-Space>", fzf.builtin)
        vim.keymap.set("n", "f<Space>", fzf.files)
        vim.keymap.set("n", "s<Space>", module_api_search)
        vim.keymap.set("n", "h<Space>", function()
                fzf.help_tags({ previewer = false })
        end)
        vim.keymap.set("n", "m<Space>", function()
                fzf.manpages({ previewer = false })
        end)
end
