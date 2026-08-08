-- lua/my-plugin/init.lua
local M = {}

M.config = {
  greeting = "hello from iceclimber",
}

function M.setup(opts) M.config = vim.tbl_deep_extend("force", M.config, opts or {}) end

function M.start()
  require("iceclimber.job").start()
  require("iceclimber.ui_lockdown").enable()
  require("iceclimber.socket").connect("127.0.0.1", 4545, function()
    require("iceclimber.watcher").start(
      function(state) require("iceclimber.socket").send(state) end
    )
  end)
end

function M.stop()
  require("iceclimber.job").stop()
  require("iceclimber.ui_lockdown").disable()
end

function M.say_hello() vim.notify(M.config.greeting) end

return M
