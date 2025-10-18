vim.keymap.set("n", "en", 'oif err != nil {<CR>}<ESC>O',                         { noremap = true, silent = true })
vim.keymap.set("n", "EN", 'Iif <ESC>mzaerr := <ESC>A; err != nil {<CR>}<ESC>`z', { noremap = true, silent = true })
