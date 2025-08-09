return {
	"stevearc/oil.nvim",
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
}
