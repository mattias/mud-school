-- Scripts in this folder are loaded when the server starts and reloaded when
-- you save them (lesson 32). Start the server with:
--   ./mudschool run server -scripts scripts -admin YourName

mud.command("wave", function(player, args)
  mud.send(player, "You wave" .. (args ~= "" and (" at " .. args) or "") .. ".")
end)
