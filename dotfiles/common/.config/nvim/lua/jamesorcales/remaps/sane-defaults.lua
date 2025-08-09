local override = function(modes, new, default, desc, custom_behavior)
	local behavior = default
	if custom_behavior then
		behavior = custom_behavior
	end

	vim.keymap.set(modes, default, "<nop>")
	vim.keymap.set(modes, new, behavior, { desc = desc })
end

-------- IMPROVED FUNCTIONALITY --------

vim.keymap.set("n", "q:", ":q", { desc = "use q? instead" })

vim.keymap.set("n", "H", "<nop>")

-- Center screen on search result
vim.keymap.set("n", "n", "nzzzv")
vim.keymap.set("n", "N", "Nzzzv")

-- Center screen on jump
override({ "n", "v" }, "<S-Down>",   "<C-D>", "Page down, center cursor"         ,  "<C-D>zz"           )
override({ "n", "v" }, "<S-Up>",     "<C-U>", "Page up, center cursor"           ,  "<C-U>zz"           )
override({ "n", "v" }, "<C-S-Down>", "<C-D>", "Page down multiple, center cursor",  "<C-D><C-D><C-D>zz" )
override({ "n", "v" }, "<C-S-Up>",   "<C-U>", "Page up multiple, center cursor"  ,  "<C-U><C-U><C-U>zz" )

-- Escape in insert mode
-- <C-c> is remapped to <Esc> so that exiting insert mode behaves consistently.
-- This is especially useful for visual block insert mode (Ctrl+v + Shift+i),
-- where <Esc> is required to apply the changes to all selected lines.
-- Without this mapping, using <C-c> instead of <Esc> will cancel the block operation.
vim.keymap.set("i", "<C-c>", "<Esc>")

--------------------------------------------------------------------------------
--------------------------------------------------------------------------------

-------- NEW FUNCTIONALITY --------

-- Move selection in visual mode
vim.keymap.set("v", "<C-Down>", ":m '>+1<CR>gv=gv")
vim.keymap.set("v", "<C-Up>", ":m '<-2<CR>gv=gv")

-- Delete without affecting default register
vim.keymap.set("v", "D", [["_d]])

-- Case-insensitive search and replace without confirmation
vim.keymap.set("n", "sr", [[:%s/\<<C-r><C-w>\>/<C-r><C-w>/gI<Left><Left><Left>]])

-- Extend gs to 100ms
vim.keymap.set("n", "gs", "<cmd>sleep 100m<CR>")

---- Yank ----
-- Yank to system clipboard
vim.keymap.set({ "n", "v" }, "<C-Y>", [["+y]])
vim.keymap.set("n", "<C-Y><C-Y>", [["+yy]])

-- Open File Explorer
vim.keymap.set("n", "-", vim.cmd.Ex)

-- Add character to end of line
vim.keymap.set("n", ",", "mzA,<ESC>`z")
