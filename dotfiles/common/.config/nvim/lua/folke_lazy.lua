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
                                vim.keymap.set("n", "<Space>", fzf.files)
                                vim.keymap.set("n", "<C-Space>", fzf.builtin)
                                vim.keymap.set("n", "s<Space>",  fzf.live_grep)
                                vim.keymap.set("n", "h<Space>", function() fzf.help_tags ({ previewer = false }) end )
                                vim.keymap.set("n", "m<Space>", function() fzf.manpages  ({ previewer = false }) end )
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
