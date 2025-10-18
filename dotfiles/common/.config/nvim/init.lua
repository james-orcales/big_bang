local palette = {
        ["yellow"]     = "#F6C177",
        ["red"]        = "#EB6F92",
        ["blue"]       = "#9CCFD8",
        ["text_dark"]  = "#777777",
}
vim.cmd.colorscheme("quiet")
vim.api.nvim_set_hl(0, "Comment",     { fg = palette["text_dark"] })
vim.api.nvim_set_hl(0, "String",      { fg = palette["yellow"]    })
vim.api.nvim_set_hl(0, "Directory",   { fg = palette["blue"]      })
vim.api.nvim_set_hl(0, "Visual",      { bg = "#333333",           })
vim.api.nvim_set_hl(0, "NormalFloat", { bg = "#0A0A0A"            })
vim.api.nvim_set_hl(0, "StatusLine",  { bg = "#111111"            })
vim.api.nvim_set_hl(0, "StatusLine",  { bg = "#111111"            })
vim.api.nvim_set_hl(0, "TODO",        { fg = palette["red"]       })
vim.api.nvim_set_hl(0, "YankSystemClipboard", { bg = "#0000FF", fg = "#FFFFFF" })


vim.api.nvim_create_autocmd("TextYankPost", {
        desc     = "Highlight when yanking text",
        group    = vim.api.nvim_create_augroup("highlight-yank", { clear = true }),
        callback = function()
                local event = vim.v.event
                local yank_to_clipboard = event.regname == "+"
                if yank_to_clipboard then
                        vim.highlight.on_yank({higroup = "YankSystemClipboard"})
                else
                        vim.highlight.on_yank()
                end
        end,
})


vim.api.nvim_create_autocmd({ "BufWritePre" }, {
        desc = "Spacing Format on save",
        group = vim.api.nvim_create_augroup("spacing-format-write", { clear = true }),
        pattern = {"*.odin", "*.go", "*.py", "*.sh", "*.lua"},
        callback = function()
                -- NOTE: I tried Ex commands with regex at first. The problem was that when undoing, the cursor would
                -- jump to the top of the file.
                --
                -- local pos = vim.api.nvim_win_get_cursor(0)
                -- -- https://vim.fandom.com/wiki/Regex_lookahead_and_lookbehind
                -- vim.cmd([[:%s/^\s*\n\{1,}/\r\r]])
                -- vim.api.nvim_win_set_cursor(0, pos)
                local bufnr        = vim.api.nvim_get_current_buf()
                local lines        = vim.api.nvim_buf_get_lines(bufnr, 0, -1, false)
                local new_lines    = {}
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
                vim.api.nvim_buf_set_lines(bufnr, 0, -1, false, new_lines)
        end,
})


require 'keymap'
require 'option'
-- prefixed with `folke` to avoid clashing with `lazy` plugin.
-- could also namespace the files by creating `nvim/lua/<NAMESPACE>/<module>.lua` and then `require '<NAMESPACE>' but I want to simplify the directory structure
require 'folke_lazy'
