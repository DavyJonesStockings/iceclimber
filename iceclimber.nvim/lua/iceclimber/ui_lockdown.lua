local M = {}

local saved = {}

function M.enable()
  saved.diagnostic = vim.diagnostic.is_enabled()
  vim.diagnostic.enable(false)

  saved.updatetime = vim.o.updatetime
  vim.o.updatetime = 10000
end

function M.disable()
  vim.diagnostic.enable(saved.diagnostic)
  vim.o.updatetime = saved.updatetime
end

return M
