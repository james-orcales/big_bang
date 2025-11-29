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

vim.api.nvim_set_hl(0, "rustCommentLineDoc",          { link = "Comment" })
-- stylua: ignore end

-- === Options ===
vim.diagnostic.config({ virtual_lines = { current_line = true } })
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

vim.opt.makeprg = ""

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

-- Center screen on search result
xplat_set("n", "n", "nzzzv")
xplat_set("n", "N", "Nzzzv")

vim.keymap.set({ "v", "n", "i" }, "<C-E>", function()
        vim.cmd("w")
        local path = vim.api.nvim_buf_get_name(0)
        if path:match("%.go$") or path:match("%.rs") or path:match("%.odin") then
                local global_mp = vim.opt.makeprg:get() or ""
                local local_mp = vim.api.nvim_get_option_value("makeprg", { scope = "local" }) or ""
                if global_mp == "" and local_mp == "" then
                        local new_makeprg = vim.fn.input("makeprg is unset. Enter command: ")
                        if new_makeprg ~= "" then
                                vim.api.nvim_set_option_value("makeprg", new_makeprg, { scope = "global" })
                        end
                end
                vim.cmd("silent mak!")
                if #vim.fn.getqflist() > 0 then
                        vim.cmd("cope")
                        vim.cmd("wincmd w")
                else
                        vim.cmd("cclo")
                end
        end
        vim.api.nvim_feedkeys(vim.api.nvim_replace_termcodes("<ESC>", true, false, true), "n", true)
end, { desc = "Save file then check" })

-- Quickfix
xplat_set("n", "{", function()
        if #vim.fn.getqflist() == 1 then
                vim.cmd("silent! cfirst")
        else
                vim.cmd("silent! cprev")
        end
end)

xplat_set("n", "}", function()
        if #vim.fn.getqflist() == 1 then
                vim.cmd("silent! cfirst")
        else
                vim.cmd("silent! cnext")
        end
end)
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
xplat_set("n", "sr", [[:%s/<C-r><C-w>/<C-r><C-w>/gI<Left><Left><Left>]])
-- Visual mode case-sensitive search and replace
xplat_set("v", "sr", [[:s///g<Left><Left><Left>]])

-- Extend gs to 100ms. Useful in macros involving lsp go-to-definition which has a little delay.
xplat_set("n", "gs", "<cmd>sleep 100m<CR>")

vim.keymap.set({ "n", "v" }, "<C-Y>", [["+y]], { desc = "Yank to system clipboard" })
vim.keymap.set("n", "<C-Y><C-Y>", [["+yy]], { desc = "Yank line to system clipboard" })

-- Add character to end of line
xplat_set("n", ",", "mzA,<ESC>`z")
xplat_set("n", ";", "mzA;<ESC>`z")

vim.keymap.set({ "v", "x" }, "gW", function()
        local original_tw = vim.opt_local.textwidth:get()
        vim.opt_local.textwidth = 100
        vim.cmd("normal! gw")
        vim.opt_local.textwidth = original_tw
end, { desc = "Temporarily wrap selection to 100 columns" })

-- === Autocmd ===

--  Workflow is `:set mp=go\ test\ ./foo` -> `:mak`. This populates quickfix with errors.
vim.api.nvim_create_autocmd({ "FileType" }, {
        desc = "Set errorformat",
        group = vim.api.nvim_create_augroup("set_errorformat", { clear = true }),
        callback = function(ev)
                if ev.match == "odin" then
                        vim.opt_local.makeprg = "vendor/Odin/odin check"
                        vim.opt_local.errorformat = "%f(%l:%v) %m"
                elseif ev.match == "go" then
                        vim.opt_local.errorformat = "%f:%l:%v: %m"
                elseif ev.match == "rust" then
                        vim.opt_local.makeprg = "cargo check"
                        vim.opt_local.errorformat = "%Eerror[E%n]: %m,%Eerror: %m,%Wwarning: %m,%Inote: %m,%C %#--> %f:%l:%c,%-G%.%#"
                        vim.opt_local.errorformat = "%-Gerror: could not compile %.%#," --ignore
                                .. "%-Gwarning: %.%# generated %.%#," --ignore
                                .. "%Eerror[E%n]: %m,%Eerror: %m,%Wwarning: %m,%Inote: %m,"
                                .. "%C %#--> %f:%l:%c,%-G%.%#"
                end
        end,
})

vim.api.nvim_create_autocmd("QuickFixCmdPost", {
        callback = function()
                local q = vim.fn.getqflist()
                table.sort(q, function(a, b)
                        return a.bufnr == b.bufnr and a.lnum < b.lnum or a.bufnr < b.bufnr
                end)
                vim.fn.setqflist(q, "r")
        end,
})

vim.api.nvim_create_autocmd({ "FileType" }, {
        desc = "Language-specific macros",
        group = vim.api.nvim_create_augroup("language_specific_macros", { clear = true }),
        callback = function(ev)
                if ev.match == "go" then
                        vim.keymap.set("n", "en", "oif err != nil {<CR>}<ESC>O", { noremap = true, silent = true, buffer = ev.buf })
                        vim.keymap.set("n", "EN", "Iif <ESC>mzaerr := <ESC>A; err != nil {<CR>}<ESC>`z", { noremap = true, silent = true, buffer = ev.buf })
                end
        end,
})

vim.api.nvim_create_autocmd({ "BufWinEnter" }, {
        desc = "Get git branch of opened buffer for statusline",
        group = vim.api.nvim_create_augroup("statusline_git_branch", { clear = true }),
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
        pattern = { "*.go", "*.lua", "*.py", "*.rs" },
        callback = function(ev)
                local lines = vim.api.nvim_buf_get_lines(ev.buf, 0, -1, false)
                local ft = vim.bo[ev.buf].filetype
                local cmd
                if ft == "lua" then
                        cmd = { "stylua", "-" }
                elseif ft == "go" then
                        cmd = { "goimports" }
                -- WHY THE FUCK WOULD NEOVIM IMPLICITLY rs -> rust
                elseif ft == "rust" then
                        cmd = { "rustfmt" }
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
                assert(#lines > 0)
                vim.api.nvim_buf_set_lines(ev.buf, 0, -1, false, lines)
        end,
})
-- Automatically formats files on save by normalizing blank lines:
--     any run of one or more blank lines becomes exactly two blank lines.
-- vim.api.nvim_create_autocmd({ "BufWritePre" }, {
--         desc = "Format on save",
--         group = format_on_save_group,
--         pattern = { "*.odin", "*.sh", "*.lua" },
--         callback = function(ev)
--                 -- NOTE: I tried Ex commands with regex at first. The problem was that when undoing, the cursor would
--                 -- jump to the top of the file.
--                 --
--                 -- local pos = vim.api.nvim_win_get_cursor(0)
--                 -- -- https://vim.fandom.com/wiki/Regex_lookahead_and_lookbehind
--                 -- vim.cmd([[:%s/^\s*\n\{1,}/\r\r]])
--                 -- vim.api.nvim_win_set_cursor(0, pos)
--                 local lines = vim.api.nvim_buf_get_lines(ev.buf, 0, -1, false)
--                 local new_lines = {}
--                 local blank_streak = 0
--                 for _, line in ipairs(lines) do
--                         if line:match("^%s*$") then
--                                 blank_streak = blank_streak + 1
--                         else
--                                 if blank_streak > 0 then
--                                         table.insert(new_lines, "")
--                                         table.insert(new_lines, "")
--                                         blank_streak = 0
--                                 end
--                                 table.insert(new_lines, line)
--                         end
--                 end
--                 vim.api.nvim_buf_set_lines(ev.buf, 0, -1, false, new_lines)
--         end,
-- })

vim.api.nvim_create_autocmd("FileType", {
        pattern = { "*" },
        group = vim.api.nvim_create_augroup("shrink_text_width", { clear = true }),
        callback = function()
                local ft = vim.bo.filetype
                local name = vim.fn.expand("%:t")
                vim.wo.colorcolumn = (ft == "markdown" or name:match("^LICENSE.*")) and "100" or "160"
                vim.bo.textwidth = (ft == "markdown" or name:match("^LICENSE.*")) and 100 or 160
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
                                        oil.set_columns({})
                                        vim.g.oil_size_column = 0
                                else
                                        oil.set_columns({ "size" })
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
end

do
        local fzf = require("fzf-lua")
        fzf.setup({
                defaults = {
                        header = false,
                },
                winopts = {
                        backdrop = 100,
                        fullscreen = true,
                        preview = { layout = "vertical", vertical = "down:50%" },
                },
                hls = { normal = "NormalFloat", border = "FloatBorder" },
                keymap = {
                        builtin = {
                                ["<S-Up>"] = "",
                                ["<S-down>"] = "",
                        },
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
        local rules = {
                Golang = {
                        Function = [[^func +(?:\([a-zA-Z0-9_]+ +\*?[a-zA-Z0-9_]+(?:\[.+\])?\))? *[A-Z][a-zA-Z0-9_]* -- !*test* ]],
                        Type = [[^type +[A-Z][a-zA-Z0-9_]+ -- !*test* ]],
                },
                Odin = {
                        Function = [[^[a-zA-Z0-9_]+ +:: +proc -- !*test* ]],
                        Type = [[^\w+ +:: +(?:struct|union|enum|distinct) -- !*test* ]],
                },
                Rust = {
                        -- We don't filter by file extension because Rust API searches often target
                        -- individual files, unlike Go or Odin, where the package system makes it
                        -- more common to search the entire directory.
                        Function_and_Macro = [[(^\s*pub (const )?(unsafe )?fn +[a-zA-Z0-9_#]+|^\s*macro_rules! [a-zA-Z0-9_#]+|^impl )]],
                        Type = [[^\s*pub (?:struct|union|enum|trait|type) [a-zA-Z0-9_#]+]],
                },
        }
        local parse_programming_language = function(path)
                if path:match("%.go$") or path == "go.mod" then
                        return "Golang"
                elseif path:match("%.odin$") then
                        return "Odin"
                elseif path:match("%.rs$") or path:lower() == "cargo.toml" then
                        return "Rust"
                end
                return nil
        end
        local module_api_search = function()
                local programming_language = nil
                local path = vim.api.nvim_buf_get_name(0)
                local operation = fzf.grep
                if not path:match("^oil://.*") then
                        programming_language = parse_programming_language(path)
                else
                        local handle = vim.uv.fs_scandir(vim.uv.cwd())
                        if handle then
                                while true do
                                        local name, t = vim.uv.fs_scandir_next(handle)
                                        if not name then
                                                break
                                        end
                                        if t == "file" then
                                                programming_language = parse_programming_language(name)
                                                if programming_language then
                                                        break
                                                end
                                        end
                                end
                        end
                end
                if programming_language == nil then
                        operation()
                        return
                else
                        if not path:match("^oil://.*") and programming_language == "Rust" then
                                operation = fzf.grep_curbuf
                        end
                        local items = {}
                        for item in pairs(rules[programming_language]) do
                                table.insert(items, item)
                        end
                        table.sort(items)
                        table.insert(items, "Any")
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
                                                else
                                                        operation({
                                                                search = rules[programming_language][selected],
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
                                                end
                                        end,
                                },
                        })
                end
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
