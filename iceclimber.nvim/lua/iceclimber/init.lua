-- lua/my-plugin/init.lua
local M = {}

M.config = {
  greeting = "hello from iceclimber",
}

function M.setup(opts)
  M.config = vim.tbl_deep_extend("force", M.config, opts or {})
end

function M.start()
  require("iceclimber.job").start()
end

function M.stop()
  require("iceclimber.job").stop()
end

function M.say_hello()
  vim.notify(M.config.greeting)
end

return M
