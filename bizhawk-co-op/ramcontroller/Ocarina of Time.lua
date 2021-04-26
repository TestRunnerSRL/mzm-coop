-- romname: OoTR

console.log('------------------')
local old_global_metatable = getmetatable(_G)
setmetatable(_G, {
	__newindex = function (_, n)
		error("Created global variable \""..n.."\".\nDidn't you want this to be local?\nIf you actually wanted a global variable,\nuse the \"declare\" function instead.", 2)
	end,
})
local function declare (name, initval)
	rawset(_G, name, initval or false)
end

local oot = require('bizhawk-co-op\\helpers\\oot')

local oot_rom = {}

local rando_context = mainmemory.read_u32_be(0x1C6E90 + 0x15D4) - 0x80000000
if (rando_context == 0) then
	setmetatable(_G, old_global_metatable)
	error("This ROM is not compatible with this version of the co-op script or the game is not fully loaded yet.")
end

local coop_context = mainmemory.read_u32_be(rando_context + 0x0000) - 0x80000000
local protocol_version_addr = coop_context + 0
local player_id_addr        = coop_context + 4
local player_name_id_addr   = coop_context + 5
local incoming_player_addr  = coop_context + 6
local incoming_item_addr    = coop_context + 8
local outgoing_key_addr     = coop_context + 12
local outgoing_item_addr    = coop_context + 16
local outgoing_player_addr  = coop_context + 18
local player_names_addr     = coop_context + 20

local save_context = 0x11A5D0
local internal_count_addr = save_context + 0x90

-- check protocol version
local script_protocol_version = 2
local rom_protocol_version = mainmemory.read_u32_be(protocol_version_addr)
if (rom_protocol_version ~= script_protocol_version) then
	setmetatable(_G, old_global_metatable)
	error("This ROM is not compatible with this version of the co-op script\nScript protocol version: "..script_protocol_version.."\nROM protocol version: "..rom_protocol_version)
end

-- get your player num
local player_num = mainmemory.read_u8(player_id_addr)

-- gives an item
local get_item = function(item)
	if (item.i == 0) then
		-- Trying to place padded items
		printOutput("[Warn] Received an invalid item!")

		-- Don't give the item but increment the internal_count
		local internal_count = mainmemory.read_u16_be(internal_count_addr)
		internal_count = internal_count + 1
		mainmemory.write_u16_be(internal_count_addr, internal_count)
		return
	end

	mainmemory.write_u16_be(incoming_player_addr, item.t)
	mainmemory.write_u16_be(incoming_item_addr, item.i) -- this is the actual item to give
end


oot_rom.itemcount = 1

local sent_items = {}
local received_items = { [0] = {f = player_num, t = 0, k = 0, i = 0} }
local received_counter = 0
local send_player_name = false
local send_player_items = false
local player_names = {}


local save_entry = function(key, value)
	if value.i == 0 then
		return
	end

	-- open file
	local file_loc = '.\\bizhawk-co-op\\savedata\\' .. gameinfo.getromname() .. '.dat'
	local f = io.open(file_loc, "a")

	if f then
		f:write(key .. ',' .. tabletostring(value) .. '\n')
		f:close()
	end
end


local load_save = function()
	-- open file
	local file_loc = '.\\bizhawk-co-op\\savedata\\' .. gameinfo.getromname() .. '.dat'
	local f = io.open(file_loc, "r")

	if f then
		sent_items = {}
		received_items = { [0] = {f = player_num, t = 0, k = 0, i = 0} }
		received_counter = 0

		for line in f:lines() do
			local splitline = strsplit(line, ',', 1)
			local key = splitline[1]
			local splitvalue = strsplit(splitline[2], ',')
			local value = stringtotable(splitvalue)

			if key == "sent" then
				table.insert(sent_items, value)
			elseif key == "received" then
				table.insert(received_items, value)
				received_counter = received_counter + 1
			elseif key == "name" then
				player_names[value["i"]] = value["n"]
			end
		end
		f:close()
	end
end
load_save()


local shop_scenes = {[0x2C]=1, [0x2D]=1, [0x2E]=1, [0x2F]=1, [0x30]=1, [0x31]=1, [0x32]=1, [0x33]=1, [0x42]=1, [0x4B]=1}
local function safeToGiveItem()
	local details
	local scene
	_, details = oot.get_current_game_mode()
	scene = oot.ctx:rawget('cur_scene'):rawget()
	return details.name == "Normal Gameplay" and shop_scenes[scene] == nil
end

local function processQueue()
	local item_id = mainmemory.read_u16_be(incoming_item_addr)
	if safeToGiveItem() and item_id == 0 then
		local internal_count = mainmemory.read_u16_be(internal_count_addr)
		-- if internal counter is ahead, we do not know what items
		-- that are missing. We will set the internal count down
		-- to what the script is aware of. The sync timer should
		-- catch this if there are indeed missing items and we should
		-- be able to sync up again. It's possible this may cause
		-- us to receive duplicated items, but this is a more stable
		-- behavior.

		if received_counter < internal_count then
			mainmemory.write_u16_be(internal_count_addr, received_counter)
			printOutput("[Warn] Game has more items than Script is aware of.")
		end
		-- if the internal counter is behind, give the next item
		if received_counter > internal_count then
			local item = received_items[internal_count + 1]
			get_item(item)
		end
	end
end


local table_has_key = function(table, key)
	for _,item in pairs(table) do
		local found = true
		for k,v in pairs(item) do
			if v ~= key[k] then
				found = false
			end
		end
		if found then
			return true
		end
	end
	return false
end


local write_name = function(id, name)
	local name_address = player_names_addr + (id * 8)
	local name_index = 0

	for _,c in pairs({string.byte(name, 1, 100)}) do
		if c >= string.byte('0') and c <= string.byte('9') then
			c = c - string.byte('0')
		elseif c >= string.byte('A') and c <= string.byte('Z') then
			c = c + 0x6A
		elseif c >= string.byte('a') and c <= string.byte('z') then
			c = c + 0x64
		elseif c == string.byte('.') then
			c = 0xEA
		elseif c == string.byte('-') then
			c = 0xE4
		elseif c == string.byte(' ') then
			c = 0xDF
		else
			c = nil
		end

		if c ~= nil then
			mainmemory.write_u8(name_address + name_index, c)

			name_index = name_index + 1
			if name_index >= 8 then
				break
			end
		end	
	end

	for i = name_index, 7 do
		mainmemory.write_u8(name_address + i, 0xDF)
	end
end


local table_count = function (table)
	local count = 0
	for _, _ in pairs(table) do
		count = count + 1
	end
	return count
end


local validate_received_items = function ()
	local new_received_items = table_count(received_items) - 1
	if new_received_items ~= received_counter then
		printOutput("[Warn] Inconsistent table size with counter. Attempting to correct...")
		received_counter = new_received_items
	end

	local item_counts = {}
	for _,item in pairs(received_items) do
		item_counts[item.f] = (item_counts[item.f] or 0) + 1
	end

	return item_counts
end
local validation_timer = coroutine.create(timer_coroutine)
coroutine.resume(validation_timer, 60, validate_received_items)


-- Gets a message to send to the other player of new changes
-- Returns the message as a dictionary object
-- Returns false if no message is to be send
function oot_rom.getMessage()
	local message = {}
	local has_content = false

	-- runs every frame
	processQueue()

	for id, name in pairs(player_names) do
		if mainmemory.read_u32_be(player_names_addr + (id * 8) + 0) == 0xDFDFDFDF and
		   mainmemory.read_u32_be(player_names_addr + (id * 8) + 4) == 0xDFDFDFDF then
			write_name(id, name)
		end
	end

	-- if there is a item pending to give to another player, make a message for it and clear it
	local key = mainmemory.read_u32_be(outgoing_key_addr)
	if key ~= 0 then
		-- create the message
		local item = mainmemory.read_u16_be(outgoing_item_addr)
		local player = mainmemory.read_u16_be(outgoing_player_addr)
		has_content = true
		message["m"] = {[0] = { f = player_num, t = player, k = key, i = item } }

		if not table_has_key(sent_items, message.m) then
			table.insert(sent_items, message.m)
			save_entry("sent", message.m[0])
		end

		-- clear the pending item data
		mainmemory.write_u32_be(outgoing_key_addr, 0)
		mainmemory.write_u16_be(outgoing_item_addr, 0)
		mainmemory.write_u16_be(outgoing_player_addr, 0)
	end

	-- send my player name if event is queued
	if send_player_name then
		has_content = true
		send_player_name = false
		message["n"] = player_num
	end

	-- resend items to queued player
	if send_player_items ~= false then
		has_content = true
		if message["m"] == nil then
			message["m"] = {}
		end

		for _,item in pairs(sent_items) do
			if item.t == send_player_items then
				table.insert(message["m"], item)
			end
		end
		send_player_items = false
	end

	-- send received item counts for validation
	local timer_status, item_counts = coroutine.resume(validation_timer)
	if item_counts then
		has_content = true
		message["c"] = {t=player_num, f=item_counts}
	end

	-- return the messages
	if has_content then
		return message
	else
		return false
	end
end


-- Process a message from another player and update RAM
function oot_rom.processMessage(their_user, message)
	-- "i" type is for handling item split events, which
	-- is not something this ram controller does. However
	-- this event will happen any time a player joins,
	-- so we will send player names again and reload the
	-- script's save data for stability
	if message["i"] then
		player_names[player_num] = config.user
		save_entry("name", {i=player_num, n=config.user})
		write_name(player_num, config.user)
		send_player_name = true
	end

	-- player name message from another player
	if message["n"] then
		player_names[message["n"]] = their_user
		save_entry("name", {i=message["n"], n=their_user})
		write_name(message["n"], their_user)
	end

	-- Player item counts from another player:
	-- We want to check for consistency with the items
	-- we've sent. If there is an inconsistency, then
	-- we'll resend all the items to them.
	if message["c"] then
		local sent_count = 0
		for _,item in pairs(sent_items) do
			if item.t == message["c"].t then
				sent_count = sent_count + 1
			end
		end

		local other_count = message["c"].f[player_num] or 0
		if other_count < sent_count then
			printOutput(their_user .. " is potentially missing items. Attempting to resync...")
			send_player_items = message["c"].t
		end
	end

	-- item get message from another player
	if message["m"] then
		if table_count(message["m"]) > 1 then
			printOutput("[Warn] " .. their_user .. " detected you were missing items. Attempting to resync...")
		end

		for _,item in pairs(message["m"]) do
			-- check if this is for this player, otherwise, ignore it
			if item.t == player_num or item.i == 0xCA then
				-- Check if this item has been received already
				if not table_has_key(received_items, item) then
					-- queue up the item get
					table.insert(received_items, item)
					received_counter = received_counter + 1
					save_entry("received", item)
				end
			end
		end
	end
end


setmetatable(_G, old_global_metatable)

return oot_rom
