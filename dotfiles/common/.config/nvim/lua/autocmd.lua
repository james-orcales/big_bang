vim.api.nvim_create_autocmd("TextYankPost", {
        desc     = "Highlight when yanking text",
        group    = vim.api.nvim_create_augroup("highlight-yank", { clear = true }),
        callback = function()
                vim.highlight.on_yank()
        end,
})


vim.api.nvim_create_autocmd({ "BufWritePre" }, {
        desc = "Spacing Format on save",
        group = vim.api.nvim_create_augroup("spacing-format-write", { clear = true }),
        pattern = {"*.odin", "*.go", "*.py"},
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
