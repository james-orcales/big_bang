return {
    'mbbill/undotree',
    config = function()
        vim.g.undotree_WindowLayout = 4
        vim.g.undotree_shortIndicators = 1
        vim.g.undotree_SetFocusWhenToggle = 1
    end,
    keys = {
        { "<leader>ut", vim.cmd.UndotreeToggle }
    }
}
