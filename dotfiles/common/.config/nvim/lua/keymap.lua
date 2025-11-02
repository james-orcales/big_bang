vim.g.mapleader = "e"


local os = vim.uv.os_uname().sysname
local set
if os == "Darwin" then
        set = function(modes, lhs, rhs, opts)
                lhs = lhs:gsub("[cC]%-", "M-") -- replace control with Option
                assert(type(lhs) == "string")
                vim.keymap.set(modes, lhs, rhs, opts)
        end
else
        set = vim.keymap.set
end


vim.keymap.set({ "n", "v", "o" }, "<C-Home>", "gg", { desc = "Jump first line" })
vim.keymap.set({ "n", "v", "o" }, "<C-End>",  "G",  { desc = "Jump last line"  })
set({ "n", "v", "o", "i"       }, "<M-b>",      "b",                 { desc = "Jump previous word - MacOS"        })
set({ "n", "v", "o", "i"       }, "<C-Left>",   "b",                 { desc = "Jump previous word"                })
set({ "n", "v", "o", "i"       }, "<S-Left>",   "B",                 { desc = "Jump previous whitespace"          })
set({ "n", "v", "o", "i"       }, "<M-f>",      "w",                 { desc = "Jump next word - MacOS"            })
set({ "n", "v", "o", "i"       }, "<C-Right>",  "w",                 { desc = "Jump next word"                    })
set({ "n", "v", "o", "i"       }, "<S-Right>",  "W",                 { desc = "Jump next whitespace"              })
set({ "n", "v", "i"            }, "<S-Down>",   "<C-D>zz",           { desc = "Page down,          center cursor" })
set({ "n", "v", "i"            }, "<C-S-Down>", "<C-D><C-D><C-D>zz", { desc = "Page down multiple, center cursor" })
set({ "n", "v", "i"            }, "<S-Up>",     "<C-U>zz",           { desc = "Page up,            center cursor" })
set({ "n", "v", "i"            }, "<C-S-Up>",   "<C-U><C-U><C-U>zz", { desc = "Page up multiple,   center cursor" })
set({ "i", "c"                 }, "<C-H>",      "<C-W>",             { desc = "Kill word before cursor"           })
set({ "i", "c"                 }, "<C-BS>",     "<C-W>",             { desc = "Kill word before cursor"           })


vim.keymap.set({ "n", "v", "o" }, "<Home>",    "^zH",       { desc = "Jump to first char of current line and screen hug left" })
vim.keymap.set({ "i"           }, "<Home>",    "<ESC>^zHi", { desc = "Jump to first char of current line and screen hug left" })


vim.keymap.set({"n", "i"}, "<C-E>",   "<ESC>:w<CR>", { desc = "Save File" })


-- Typos
set("i", "!+",   "!=")
set("i", ":+",   ":=")


-------- IMPROVED FUNCTIONALITY --------


set("n", "H", "<nop>")


-- Center screen on search result
set("n", "n", "nzzzv")
set("n", "N", "Nzzzv")


-- Escape in insert mode
-- <C-c> is remapped to <Esc> so that exiting insert mode behaves consistently.
-- This is especially useful for visual block insert mode (Ctrl+v + Shift+i),
-- where <Esc> is required to apply the changes to all selected lines.
-- Without this mapping, using <C-c> instead of <Esc> will cancel the block operation.
vim.keymap.set("i", "<C-c>", "<Esc>")


--------------------------------------------------------------------------------
--------------------------------------------------------------------------------


-------- NEW FUNCTIONALITY --------


-- Quickfix
set("n", "{", "<CMD>:cprevious<CR>")
set("n", "}", "<CMD>:cnext<CR>")


-- Move selection in visual mode
set("v", "<C-Down>", ":m '>+1<CR>gv=gv")
set("v", "<C-Up>", ":m '<-2<CR>gv=gv")


-- Delete without affecting default register
set("v", "D", [["_d]])


-- Case-insensitive search and replace without confirmation
-- TODO: Support visual mode
set("n", "sr", [[:%s/\<<C-r><C-w>\>/<C-r><C-w>/gI<Left><Left><Left>]])


-- Extend gs to 100ms. Useful in macros involving lsp go-to-definition which has a little delay.
set("n", "gs", "<cmd>sleep 100m<CR>")


vim.keymap.set({ "n", "v" }, "<C-Y>",      [["+y]],  { desc = "Yank to system clipboard" })
vim.keymap.set("n",          "<C-Y><C-Y>", [["+yy]], { desc = "Yank line to system clipboard" })


-- Open File Explorer
set("n", "-", vim.cmd.Ex)


-- Add character to end of line
set("n", ",", "mzA,<ESC>`z")
