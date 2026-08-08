local M = {}

local saved_win_opts = {}

local function enforce_platformer_view(win)
  saved_win_opts[win] = {
    wrap = vim.wo[win].wrap,
  }
  vim.wo[win].wrap = false
end

local function restore_view(win)
  local saved = saved_win_opts[win]
  if saved then
    vim.wo[win].wrap = saved.wrap
    saved_win_opts[win] = nil
  end
end

local function get_visible_state()
  local win = vim.api.nvim_get_current_win()
  local buf = vim.api.nvim_get_current_buf()
  local top = vim.fn.line("w0")
  local bot = vim.fn.line("w$")
  local raw_lines = vim.api.nvim_buf_get_lines(buf, top - 1, bot, false)
  local view = vim.fn.winsaveview()

  local lines = {}
  for i, text in ipairs(raw_lines) do
    lines[i] = {
      text = text,
      width = vim.fn.strdisplaywidth(text),
    }
  end

  return {
    win = win,
    buf = buf,
    top = top,
    bot = bot,
    win_width = vim.api.nvim_win_get_width(win),
    win_height = vim.api.nvim_win_get_height(win),
    leftcol = view.leftcol,
    lines = lines,
    cursor = vim.api.nvim_win_get_cursor(win),
  }
end

function M.start(on_update)
  local win = vim.api.nvim_get_current_win()
  enforce_platformer_view(win)

  local group = vim.api.nvim_create_augroup("IceClimbeWatcher", { clear = true })

  vim.api.nvim_create_autocmd({ "WinScrolled", "VimResized", "WinNew", "WinClosed" }, {
    group = group,
    callback = function()
      on_update(get_visible_state())
    end,
  })

  vim.api.nvim_create_autocmd("BufEnter", {
    group = group,
    callback = function(args)
      vim.api.nvim_buf_attach(args.buf, false, {
        on_lines = function()
          vim.schedule(function()
            on_update(get_visible_state())
          end)
        end,
      })
    end,
  })

  on_update(get_visible_state())
end

function M.stop()
  vim.api.nvim_clear_autocmds({ group = "IceClimbeWatcher" })
  for win, _ in pairs(saved_win_opts) do
    if vim.api.nvim_win_is_valid(win) then restore_view(win) end
  end
end

return M
