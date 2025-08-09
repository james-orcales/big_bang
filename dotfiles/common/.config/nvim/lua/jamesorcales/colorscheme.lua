local palette = {
        ["yellow"]     = "#F6C177",
        ["green"]      = "#31748F",
        ["red"]        = "#EB6F92",
        ["blue"]       = "#9CCFD8",
        ["text_dark"]  = "#777777",
        ["text_mid"]   = "#AAAAAA",
        ["text_light"] = "#FFFFFF",
}
vim.cmd.colorscheme("quiet")

-- Matches strings enclosed in either double quotes or backticks, correctly handling escaped delimiters (e.g., "foo \"bar\"" or `command \`arg\``). \v sets
-- "very magic" mode for simpler regex syntax. The pattern `("..."|`...`)` uses alternation to match either type of string. Inside each string type,
-- `([^D\\]|\\.)*` matches non-delimiter/non-backslash characters or any escaped character (where 'D' is the delimiter, " or `).
-- vim.api.nvim_create_autocmd("BufEnter", {
--         callback = function()
--                 vim.cmd [[syntax match String /\v("([^"\\]|\\.)*"|`([^`\\]|\\.)*`)/]]
--         end,
-- })


vim.api.nvim_set_hl(0, "Keyword",                    { fg   = palette["green"]     })
vim.api.nvim_set_hl(0, "@keyword",                   { link = "Keyword"            })
vim.api.nvim_set_hl(0, "@keyword.conditional.lua",   { link = "Keyword"            })
vim.api.nvim_set_hl(0, "@keyword.repeat.lua",        { link = "Keyword"            })
vim.api.nvim_set_hl(0, "@keyword.repeat.lua",        { link = "Keyword"            })
vim.api.nvim_set_hl(0, "fishNot",                    { link = "Keyword"            })
vim.api.nvim_set_hl(0, "fishKeywordAndOr",           { link = "Keyword"            })
vim.api.nvim_set_hl(0, "Comment",                    { fg   = palette["text_dark"] })
vim.api.nvim_set_hl(0, "String",                     { fg   = palette["yellow"]    })
vim.api.nvim_set_hl(0, "Directory",                  { fg   = palette["blue"]      })
vim.api.nvim_set_hl(0, "Visual",                     { bg   = "#333333",           })
vim.api.nvim_set_hl(0, "NormalFloat",                { bg   = "#0A0A0A"            })
vim.api.nvim_set_hl(0, "StatusLine",                 { bg   = "#111111"            })
vim.api.nvim_set_hl(0, "StatusLine",                 { bg   = "#111111"            })
vim.api.nvim_set_hl(0, "TODO",                       { fg = palette["red"]         })
-- gray out brackets. modify treesitter highlighting since lua treesitter is bundled with nvim already
