local lazypath = vim.fn.stdpath("data") .. "/lazy/lazy.nvim"
if not (vim.uv or vim.loop).fs_stat(lazypath) then
    vim.fn.system({
        "git",
        "clone",
        "--filter=blob:none",
        "https://github.com/folke/lazy.nvim.git",
        "--branch=stable",
        lazypath,
    })
end
vim.opt.rtp:prepend(lazypath)


require("lazy").setup({ 
        spec = {
                {
                        "ggandor/leap.nvim",
                        event = "VeryLazy",
                        config = function()
                                local leap = require("leap")

                                leap.opts.safe_labels = {}
                                leap.opts.labels = "setnriaofuplwyqjbmghdzxc"
                                leap.opts.max_phase_one_targets = 0
                                leap.opts.special_keys.next_group = "<space>"

                                vim.keymap.set({ "n", "x", "o" }, "t", "<Plug>(leap)")
                                vim.api.nvim_set_hl(0, "LeapBackdrop", { link = "Comment" })
                        end,
                },
                {
                        "ibhagwan/fzf-lua",
                        pin = true,
                        event = "VimEnter",
                        config = function()
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
                                                        ["alt-i"]  = fzf.actions.toggle_ignore,
                                                        ["enter"]  = nil,
                                                        ["ctrl-s"] = nil,
                                                        ["ctrl-v"] = nil,
                                                        ["ctrl-t"] = nil,
                                                        ["alt-q"]  = nil,
                                                        ["alt-Q"]  = nil,
                                                        ["alt-h"]  = nil,
                                                        ["alt-f"]  = nil,
                                                }
                                        },
                                })
                                
                                local function module_api_search()
                                        programming_language = nil
                                        local handle = vim.uv.fs_scandir(vim.uv.cwd())
                                        if handle then
                                                while true do
                                                        local name, t = vim.uv.fs_scandir_next(handle)
                                                        if name == nil then 
                                                                break 
                                                        end
                                                        if t == "file" then
                                                                if name:match("%.go$") then
                                                                        programming_language = "Golang"
                                                                        break
                                                                end
                                                                if name:match("%.odin$") then
                                                                        programming_language = "Odin"
                                                                        break
                                                                end
                                                        end
                                                end
                                        end
                                        fzf.live_grep()
                                        if programming_language == nil then
                                                fzf.live_grep()
                                                return
                                        end

                                        local items = nil
                                        if programming_language == "Golang" then
                                                items = { "Function", "Type", "Variables", "_Function", "_Type"}
                                        elseif programming_language == "Odin" then
                                                items = { "Procedure", "Type", "Variables", "_Procedure", "_Type"}
                                        end
                                        assert(items ~= nil)
                                        table.insert(items, "Any")
                                        fzf.fzf_exec(
                                                items,
                                                {
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
                                                                        assert(programming_language ~= nil)
                                                                        local pattern = nil
                                                                        local rg_opts = nil
                                                                        if programming_language == "Golang" then
                                                                                if selected == "Function" then
                                                                                        local func       = [[^func +]]
                                                                                        local receiver   = [[(?:\(\w+ +\*?\w+\))? *]] -- optional
                                                                                        local identifier = [[[A-Z]\w+]]
                                                                                        local generics   = [[(?:\[.*?\])?]]
                                                                                        local signature  = [[\(.*?\) +]]
                                                                                        pattern = func .. receiver .. identifier .. generics .. signature
                                                                                elseif selected == "Type" then
                                                                                        pattern = [[^type +[A-Z]\w* +]]
                                                                                elseif selected == "Variables" then
                                                                                        rg_opts = "--multiline"
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
                                                                                if selected == "Procedure"  then
                                                                                        pattern = [[^\w+ +:: +proc]]
                                                                                elseif selected == "Type" then
                                                                                        pattern = [[^\w+ +:: +(?:struct|union|enum|distinct)]]
                                                                                end
                                                                                pattern = pattern .. " -- !*test*"
                                                                        end
                                                                        fzf.live_grep({ 
                                                                                search = pattern, 
                                                                                rg_opts = rg_opts,
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
                                                }
                                        )
                                end
                                vim.keymap.set("n", "<Space>",   fzf.files)
                                vim.keymap.set("n", "<C-Space>", fzf.builtin)
                                vim.keymap.set("n", "s<Space>",  module_api_search)
                                vim.keymap.set("n", "h<Space>",  function() fzf.help_tags ({ previewer = false }) end )
                                vim.keymap.set("n", "m<Space>",  function() fzf.manpages  ({ previewer = false }) end )

                        end,
                },
                {
                        "stevearc/oil.nvim",
                        pin = true,
                        config = function()
                                local oil = require("oil")
                                oil.setup({
                                        -- default_file_explorer = true,
                                        columns     = { "icon" },
                                        buf_options = { buflisted = false, bufhidden = "hide" },
                                        win_options = {
                                                wrap       = false,
                                                spell      = false,
                                                list       = false,
                                                foldcolumn = "0",
                                        },
                                        delete_to_trash                 = false,
                                        prompt_save_on_select_new_entry = true,
                                        constrain_cursor                = "name",
                                        keymaps = {
                                                ["?"]        = "actions.show_help",
                                                ["<CR>"]     = "actions.select",
                                                ["<C-C>"]    = oil.discard_all_changes,
                                                ["-"]        = "actions.parent", -- dash
                                                ["_"]        = "actions.open_cwd", -- underscore
                                                ["cd"]        = "actions.cd",
                                                ["<C-Home>"] = "gg",
                                                ["<C-End>"]  = "G",
                                                ["="]        = function()
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
                        end,
                },
                {
                        "kylechui/nvim-surround",
                        pin = true,
                        version = "3.1.3",
                        commit  = "7a7a78a52219a3312c1fcabf880cea07a7956a5f",
                        pin = true,
                        event   = "VeryLazy",
                        opts = {
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
                        },
                },
                {
                        "junegunn/vim-easy-align",
                        pin = true,
                        commit = "9815a55dbcd817784458df7a18acacc6f82b1241",
                        pin = true,
                        config = function()
                                vim.g.easy_align_ignore_groups = {}
                                vim.keymap.set("n", "ga", "<Plug>(EasyAlign)")
                                vim.keymap.set("x", "ga", "<Plug>(EasyAlign)")
                        end,
                },
                {
                        'mbbill/undotree',
                        pin = true,
                        config = function()
                                vim.g.undotree_WindowLayout = 4
                                vim.g.undotree_shortIndicators = 1
                                vim.g.undotree_SetFocusWhenToggle = 1
                        end,
                        keys = {
                                { "<leader>ut", vim.cmd.UndotreeToggle }
                        }
                },
        },
})
