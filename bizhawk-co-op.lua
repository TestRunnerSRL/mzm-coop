local guiClick = {}

mainform = nil
local text1, lblRooms, btnGetRooms, ddRooms, btnQuit, btnJoin, btnHost
local txtUser, txtPass, lblUser, lblPass, ddRamCode, lblRamCode
local lblPort, txtPort
config = {}

formPlayerNumber = nil
formPlayerCount = nil
formPlayerList = nil

kicked = false

function strsplit(inputstr, sep, max)
	if not inputstr then
		return {}
	end

	if not sep then
		sep = ","
	end
	local t={} ; local i=1
	for str in string.gmatch(inputstr, "([^"..sep.."]+)") do
		if max and i > max then
			if t[i] then
				t[i] = t[i] .. sep .. str
			else
				t[i] = str
			end
		else
			t[i] = str
			i = i + 1
		end
	end
	return t
end

-- A simple function to count a table, even though some functions
-- do exist for counting, they do not work in all cases
function getTableSize(t)
	local count = 0
	for _, __ in pairs(t) do
		count = count + 1
	end
	return count
end

-- Checks if a single dimension table has a value
-- Returns true if found, false if not found
function tableHasValue(tbl, val)
	for _, v in pairs(tbl) do
		if (tostring(v) == tostring(val)) then
			return true
		end

		if(tonumber(v) == tonumber(val)) then
			return true
		end
	end

	return false
end

-- Returns a table of input tables keys sorted by the values
-- of the input table. Accepts a sort function for sort ordering
function getKeysSortedByValue(tbl, sortFunction)
	local keys = {}
	for key in pairs(tbl) do
		table.insert(keys, key)
	end

	table.sort(keys, function(a, b)
		return sortFunction(tbl[a], tbl[b])
	end)

	return keys
end

-- Returns an inverted version of the table provided
function invert_table(t)
	local inverted = {}
	for key,val in pairs(t) do
		inverted[val] = key
	end
	return inverted
end

local sync = require("bizhawk-co-op\\sync")

--Add a line to the output. Inserts a timestamp to the string
function printOutput(str)
	local text = forms.gettext(text1)
	local pos = #text
	forms.setproperty(text1, "SelectionStart", pos)

	str = string.gsub (str, "\n", "\r\n")
	str = "[" .. os.date("%H:%M:%S", os.time()) .. "] " .. str
	if pos > 0 then
		str = "\r\n" .. str
	end

	forms.setproperty(text1, "SelectedText", str)
end

--Repeatedly tries to reconnect with the host using exponential backoff.
function reconnectToHost()
	local retry = 0
	while host.reconnecting() do
		local init = os.time()
		local backoff = math.random(5, math.min(60, 5 * math.pow(2, retry)))
		printOutput("Waiting " .. tostring(backoff) .. "s to reconnect.")
		while os.difftime(os.time(), init) < backoff do
			coroutine.yield()
			if not host.reconnecting() then
				return
			end
		end
		--pcall doesn't work with functions that call coroutine.yield(), so we
		--use a coroutine instead.
		local thread = coroutine.create(host.join)
		while coroutine.status(thread) == 'suspended' do
		  coroutine.resume(thread)
		  coroutine.yield()
		end
		retry = retry + 1
	end
end


host = require("bizhawk-co-op\\host")


local roomlist = false
function refreshRooms()
	roomlist = host.getRooms()
	if roomlist then
		roomlist['(Custom IP)']='(Custom IP)'
		forms.setdropdownitems(ddRooms, roomlist)
	else
		forms.setdropdownitems(ddRooms,
				{
					['(Custom IP)']='(Custom IP)'
				})
	end

	updateGUI()
end


--Reloads all the info on the form. Disables any inappropriate components
function updateGUI()
	if host.status == 'Idle' then
		if forms.setdropdownitems and roomlist then
			forms.setproperty(ddRooms, 'Enabled', true)
		else
			forms.setproperty(ddRooms, 'Enabled', false)
		end
		forms.setproperty(btnGetRooms, "Enabled", true)
		forms.setproperty(ddRamCode, "Enabled", true)
		forms.setproperty(txtUser, "Enabled", true)
		forms.setproperty(txtPass, "Enabled", true)
		forms.setproperty(formPlayerNumber, "Enabled", true)
		forms.setproperty(txtPort, "Enabled", true)
		forms.setproperty(btnQuit, "Enabled", false)
		forms.setproperty(btnJoin, "Enabled", true)
		forms.setproperty(btnHost, "Enabled", true)
		forms.settext(btnHost, "Create Room")
		forms.setproperty(btnHost, "Enabled", true)

		if forms.gettext(ddRooms) == '(Custom IP)' then
			forms.setproperty(txtIP, "Enabled", true)
		else
			forms.setproperty(txtIP, "Enabled", false)
		end
	else
		forms.setproperty(btnGetRooms, "Enabled", false)
		forms.setproperty(ddRamCode, "Enabled", false)
		forms.setproperty(ddRooms, "Enabled", false)
		forms.setproperty(txtUser, "Enabled", false)
		forms.setproperty(txtPass, "Enabled", false)
		forms.setproperty(formPlayerNumber, "Enabled", false)
		forms.setproperty(txtPort, "Enabled", false)
		forms.setproperty(btnQuit, "Enabled", true)
		forms.setproperty(btnJoin, "Enabled", false)
		forms.setproperty(btnHost, "Enabled", true)
		if host.locked then
			forms.settext(btnHost, "Unlock Room")
		else
			forms.settext(btnHost, "Lock Room")
		end
		forms.setproperty(txtIP, "Enabled", false)
	end
end


--If the script ends, makes sure the sockets and form are closed
event.onexit(function () host.close(); forms.destroy(mainform) end)


--Load the changes from the form and disable any appropriate components
function prepareConnection()
	if roomlist then
		config.room = forms.gettext(ddRooms)
	else
		config.room = '(Custom IP)'
	end
	config.ramcode = forms.gettext(ddRamCode)
	config.user = forms.gettext(txtUser)
	config.pass = forms.gettext(txtPass)
	config.port = forms.gettext(txtPort)
	config.hostname = forms.gettext(txtIP)
end


--Quit/Disconnect click handle for the quit button
function leaveRoom()
	if (host.connected()) then
		sendMessage["Quit"] = true
	else
		host.close()
	end
end


--Returns a list of files in a given directory in case-insensitive sorted order
function os.dir(dir)
	local files = {}
	local f = assert(io.popen('dir \"' .. dir .. '\" /b /on', 'r'))
	for file in f:lines() do
		table.insert(files, file)
	end
	f:close()
	return files
end


--Returns a list of ramcontrollers and the 1-based index of the ramcontroller
--that matches the current ROM. Returns a nil index if none match.
--
--Ramcontrollers may start with the comment '-- romname: <pattern>' to indicate
--what romnames the ramcontroller supports. '<pattern>' should be a valid Lua
--pattern, and is partially matched agains the romname.
function getRamControllers(dir)
	local prefix = '-- romname: '
	local romname = gameinfo.getromname()
	local files = os.dir(dir)
	for i, file in ipairs(files) do
		local f = assert(io.open(dir .. '\\' .. file))
		local line = f:read('*l')
		f:close()
		if line:sub(1, #prefix) == prefix and string.match(romname, line:sub(#prefix + 1)) then
			return files, i
		end
	end
	return files, nil
end


math.randomseed(os.time())

--Create the form
mainform = forms.newform(470, 375, "Bizhawk Co-op")

text1 = forms.textbox(mainform, "", 263, 105, nil, 16, 225, true, true, 'Vertical')
forms.setproperty(text1, "ReadOnly", true)
--forms.setproperty(text1, "MaxLength", 32767)


if forms.setdropdownitems then -- can't update list prior to bizhawk 1.12.0
	btnGetRooms = forms.button(mainform, "Refresh Rooms", refreshRooms, 220, 10, 60, 23)
	ddRooms = forms.dropdown(mainform, {['(Fetching rooms...)']='(Fetching rooms...)'}, 80, 11, 135, 20)
	forms.setproperty(ddRooms, 'Enabled', false)
	guiClick["Refresh Rooms"] = refreshRooms;
else
	btnGetRooms = forms.button(mainform, "", function() end, 15, 10, 60, 23)
	forms.setproperty(btnGetRooms, 'Enabled', false)

	roomlist = host.getRooms()
	if roomlist then
		ddRooms = forms.dropdown(mainform, roomlist, 80, 11, 200, 20)
		forms.setproperty(ddRooms, 'Enabled', true)
	else
		ddRooms = forms.dropdown(mainform, {['(Custom IP)']='(Custom IP)'}, 80, 11, 200, 20)
		forms.setproperty(ddRooms, 'Enabled', false)
	end
end
lblRooms = forms.label(mainform, "Rooms:", 34, 13)

txtIP = forms.textbox(mainform, "", 200, 20, nil, 80, 40, false, false)
txtUser = forms.textbox(mainform, "", 200, 20, nil, 80, 64, false, false)
txtPass = forms.textbox(mainform, "", 200, 20, nil, 80, 88, false, false)
formPlayerNumber = forms.textbox(mainform, "", 200, 20, nil, 80, 114, false, false)
txtPort = forms.textbox(mainform, '50000', 200, 20, nil, 80, 138, false, false)
local ramCodes, defaultRamCodeIndex = getRamControllers("bizhawk-co-op\\ramcontroller")
ddRamCode = forms.dropdown(mainform, ramCodes, 80, 162, 200, 10)
if defaultRamCodeIndex then
	-- SelectedIndex is 0-based.
	forms.setproperty(ddRamCode, "SelectedIndex", defaultRamCodeIndex - 1)
end
lblIP = forms.label(mainform, "Host IP:", 32, 42)
lblUser = forms.label(mainform, "Username:", 19, 66)
lblPass = forms.label(mainform, "Password:", 21, 90)
forms.label(mainform, "Player #:", 29, 115)
lblPort = forms.label(mainform, "Port:", 48, 140)
lblRamCode = forms.label(mainform, "Game Script:", 10, 165)
forms.setproperty(txtPass, 'PasswordChar', '*')

btnQuit = forms.button(mainform, "Leave Room", leaveRoom, 15, 190, 85, 25)
forms.setproperty(btnQuit, 'Enabled', false)

btnHost = forms.button(mainform, "Create Room",
		function() prepareConnection(); guiClick["Host Server"] = host.start end,
		105, 190, 85, 25)

btnJoin = forms.button(mainform, "Join Room",
		function() prepareConnection(); guiClick["Join Server"] = host.join end,
		195, 190, 85, 25)


forms.label(mainform, "Players:", 288, 10, 44, 15)
formPlayerCount = forms.label(mainform, "...", 329, 10, 40, 15)
forms.label(mainform, "Ready:", 288, 25, 42, 15)
formReadyCount = forms.label(mainform, "...", 329, 25, 38, 15)
readyToggle = forms.button(mainform, "Ready", sync.readyToggle, 370, 11, 80, 23)
forms.setproperty(readyToggle, "Enabled", false)
formPlayerList = forms.textbox(mainform, "", 155, 288, nil, 293, 42, true, true, 'Vertical')
forms.setproperty(formPlayerList, "ReadOnly", true)

sendMessage = {}
local thread
local reconnectThread

updateGUI()

local threads = {}

emu.yield()
emu.yield()

---------------------
--    Main loop    --
---------------------
while 1 do
	--End script if form is closed
	if forms.gettext(mainform) == "" then
		return
	end

	if (forms.gettext(ddRooms) ~= prevRoom) then
		prevRoom = forms.gettext(ddRooms)
		updateGUI()
	end

	host.listen()

	--Create threads for the function requests from the form
	for k,v in pairs(guiClick) do
		threads[coroutine.create(v)] = k
	end
	guiClick = {}

	--Run the threads
	for k,v in pairs(threads) do
		if coroutine.status(k) == "dead" then
			threads[k] = nil
		else
			local status, err = coroutine.resume(k)
			if (status == false) then
				if (err ~= nil) then
					printOutput("Error during " .. v .. ": " .. err)
				else
					printOutput("Error during " .. v .. ": No error message")
				end
			end
		end
	end

	--If connected, run the syncinputs thread
	if host.connected() or host.reconnecting() then
		--If the thread didn't yield, then create a new one
		if thread == nil or coroutine.status(thread) == "dead" then
			thread = coroutine.create(sync.syncRAM)
		end
		local status, err = coroutine.resume(thread)

		if (status == false and err ~= nil) then
			if not kicked then
				printOutput("Error during sync: " .. tostring(err))
			else
				kicked = false
			end
		end
	end

	--If we're reconnecting, run the reconnect thread. host.reconnecting() can
	--temporarily become false during host.join(), so we continue running the
	--thread until it exits.
	if host.reconnecting() then
		--If the thread didn't yield, create a new one
		if reconnectThread == nil or coroutine.status(reconnectThread) == "dead" then
			reconnectThread = coroutine.create(reconnectToHost)
		end
	end
	if reconnectThread ~= nil and coroutine.status(reconnectThread) ~= "dead" then
		local status, err = coroutine.resume(reconnectThread)
		if (status == false and err ~= nil) then
		  printOutput("Error during reconnect: " .. tostring(err))
		end
	end

	-- 2 Emu Yields = 1 Frame Advance
	--If game is paused, then yield will not frame advance
	emu.yield()
	emu.yield()
end