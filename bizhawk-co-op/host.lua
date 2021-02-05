--author: TheOnlyOne

local socket = require("socket")
local http = require("socket.http")
local sync = require("bizhawk-co-op\\sync")

local host = {}

local server = nil
host.clients = {}
host.users = {}
host.playerlist = {}
host.playerstatus = {}
host.status = 'Idle'
host.locked = false
host.hostname = nil
host.client_ping = {}
local itemcount

function host.start()
	if (host.status == 'Host') then
		if host.locked then
			local roomstr, err = http.request('https://us-central1-mzm-coop.cloudfunctions.net/create' ..
					'?user=' .. config.user ..
					'&pass=' .. config.pass)
			if (err == 200) then
				host.locked = false
				updateGUI()
				printOutput('Room unlocked.')
			else
				printOutput('Error unlocking room [Code: ' .. (err or '') .. ']')
			end
		else
			local roomstr, err = http.request('https://us-central1-mzm-coop.cloudfunctions.net/destroy' ..
					'?user=' .. config.user ..
					'&pass=' .. config.pass)
			if (err == 200) then
				host.locked = true
				updateGUI()
				printOutput('Room locked.')
			else
				printOutput('Error locking room [Code: ' .. (err or '') .. ']')
			end
		end

		return
	end

	local ramcontroller = sync.loadramcontroller()
	if (ramcontroller == false) then
		return
	end

	if ramcontroller.getConfig then
		config.ramconfig = ramcontroller.getConfig()
		if (config.ramconfig == false) then
			return
		end
	else
		config.ramconfig = {}
	end

	itemcount = ramcontroller.itemcount
	if (itemcount == false) then
		return
	end

	if not config.user or config.user == '' then
		printOutput('Set your username before creating a room.')
		return false
	end

	if forms.gettext(formPlayerNumber) ~= ''
	and (tonumber(forms.gettext(formPlayerNumber)) == nil
	or tonumber(forms.gettext(formPlayerNumber)) <= 0)
	then
		printOutput('Player Number must be above 0 or empty.')
		return false
	end

	if not config.port or config.port == '' then
		config.port = 50000
	end

	host.status = 'Host'
	host.locked = true
	updateGUI()

	coroutine.yield()

	--create the server
	server = socket.bind("*", config.port, 1)
	if (server == nil) then
		printOutput("Error creating server. Port is probably in use.")
		host.status = 'Idle'
		host.locked = false
		updateGUI()
		return false
	end

	local ip, setport = server:getsockname()
	server:settimeout(0) -- non-blocking
	printOutput("Created server on port " .. setport)
	printOutput('Dedicated server options:' ..
			' --itemcount=' .. itemcount ..
			' --ramconfig=' .. tabletostring(config.ramconfig))

	local roomstr, err = http.request('https://us-central1-mzm-coop.cloudfunctions.net/create' ..
			'?user=' .. config.user ..
			'&pass=' .. config.pass)
	if (err == 200) then
		printOutput('Room initialized.')
	else
		printOutput('Error creating room [Code: ' .. (err or '') .. ']')
		--host.close()
		--return false
	end

	if (forms.gettext(formPlayerNumber) == '') then
		forms.settext(formPlayerNumber, "1")
	end

	host.locked = false
	host.users[config.user] = 1
	host.playerlist[config.user] = {['num'] = tonumber(forms.gettext(formPlayerNumber)), ['status'] = "Unready"}

	forms.settext(formPlayerCount, "1")
	forms.setproperty(formPlayerList, 'SelectionStart', 0)
	forms.setproperty(formPlayerList, "SelectedText", "[ ] P"..tonumber(forms.gettext(formPlayerNumber))..": "..config.user.."\r\n")
	forms.settext(formReadyCount, "0")
	forms.setproperty(readyToggle, "Enabled", true)

	hostControls = forms.newform(300, 88, "Host Controls")
	kickPlayerLabel = forms.label(hostControls, "Kick Player", 4, 5, 290, 13)
	kickPlayerSelect = forms.dropdown(hostControls, {['(Kick Player...)']='(Kick Player...)'}, 5, 20, 200, 5)
	kickPlayerBtn = forms.button(hostControls, "Kick", sync.kickPlayer, 208, 19, 72, 23)
	forms.setdropdownitems(kickPlayerSelect, invert_table(host.users))

	updateGUI()
	return true
end


function host.listen()
	if (server == nil or host.locked) then
		return false
	end

	--wait for the connection from the client
	local client, err = server:accept()

	--end execution if a client does not connect in time
	if (client == nil) then
		if err ~= "timeout" then
			printOutput("Server error: ", err)
			host.close()
		end
		return false
	end

	local clientID = 2
	while (true) do
		if (host.clients[clientID] == nil) then
			break
		end
		clientID = clientID + 1
	end

	--display the client's information
	printOutput("Player " .. clientID .. " connecting...")

	-- make sure we don't block forever waiting for input
	client:settimeout(0)
	client:setoption('linger', {['on']=false, ['timeout']=0})

	--sync the gameplay. use a coroutine because sync.syncconfig calls
	--coroutine.yield(), which doesn't work properly with pcall.
	local success, their_user
	local co = coroutine.create(sync.syncconfig)
	repeat
		success, their_user = coroutine.resume(co, client, clientID)
		coroutine.yield()
	until coroutine.status(co) ~= "suspended"
	if success and their_user then
		host.clients[clientID] = client
		host.users[their_user] = clientID
		host.client_ping[clientID] = 4
	else
		if not success then
			printOutput("Error in Listen: " .. their_user)
		end
		client:close()
		return false
	end

	printOutput(their_user .. " connected.")

	-- send item lists
	local clientMap = {[0]=1}
	local clientCount = 1
	for id,_ in pairs(host.clients) do
		clientMap[clientCount] = id
		clientCount = clientCount + 1
	end

	local itemlist = {}
	math.random(itemcount)
	for i=0,(itemcount-1) do
		itemlist[i] = clientMap[i % clientCount]
	end

	for i=1,(itemcount-1) do
		j = math.random(i + 1) - 1 -- 0 to i inclusive
		itemlist[i], itemlist[j] = itemlist[j], itemlist[i]
	end

	sync.sendItems(itemlist)
	sync.sendPlayerList(host.playerlist)
	sync.updatePlayerList(host.playerlist)
	host.updateHostControls()
	updateGUI()
	return clientID
end


function host.join()
	if (sync.loadramcontroller() == false) then
		return
	end

	if not config.room or config.room == '' then
		printOutput('Select a room to join.')
		return false
	end

	if config.room == '(Custom IP)' and (not config.hostname or config.hostname == '') then
		printOutput('Enter the IP to join.')
		return false
	end

	if not config.user or config.user == '' then
		printOutput('Set your username before joining a room.')
		return false
	end

	if not config.port or config.port == '' then
		config.port = 50000
	end

	if forms.gettext(formPlayerNumber) ~= ''
	and (tonumber(forms.gettext(formPlayerNumber)) == nil
	or tonumber(forms.gettext(formPlayerNumber)) <= 0)
	then
		printOutput('Player Number must be above 0 or empty.')
		return false
	end

	local reconnect = host.reconnecting()
	kicked = false
	host.close(reconnect)
	host.status = 'Join'
	host.locked = true
	updateGUI()

	coroutine.yield()

	if not reconnect and config.room ~= '(Custom IP)' then
		local err
		config.hostname, err = http.request('https://us-central1-mzm-coop.cloudfunctions.net/join' ..
				'?user=' .. config.room ..
				'&pass=' .. config.pass)
		if (err == 200) then
			printOutput('Joining ' .. config.room)
		else
			printOutput('Error joining room [Code: ' .. (err or '') .. ']')
			host.status = 'Idle'
			host.locked = false
			updateGUI()
			return
		end
	end

	--use a non-blocking connection so that frames can continue to be drawn
	--while attempting to connect. DNS lookups still occur synchronously, so
	--"host not found" errors will skip the while loop.
	local client = socket.tcp()
	client:settimeout(0)
	client:setoption('linger', {['on']=false, ['timeout']=0})
	local _, err = client:connect(config.hostname, config.port)
	local init = os.time()
	while err == 'timeout' and os.difftime(os.time(), init) < 2 do
		coroutine.yield()
		err = socket.skip(2, socket.select({client}, {client}, 0))
	end
	if err ~= nil then
		printOutput("Connection failed: " .. err)
		client:close()
		if reconnect then
			host.status = 'Reconnecting'
		else
			host.status = 'Idle'
			host.locked = false
		end
		updateGUI()
		return
	end

	--display the server's information
	local peername, peerport = client:getpeername()
	printOutput("Joined room " .. config.room)

	--sync the gameplay
	if (sync.syncconfig(client, nil)) then
		host.clients[1] = client
		host.client_ping[1] = 4
		host.users[host.hostname] = 1
	else
		client:close()
		host.close(reconnect)
		updateGUI()
	end

	forms.setproperty(readyToggle, "Enabled", true)

	if reconnect then
		gui.addmessage("Reconnected.")
		--rather than implement a separate method to send our status, just
		--toggle twice.
		sync.readyToggle()
		sync.readyToggle()
	end

	return
end


--when the script finishes, make sure to close the connection
function host.close(reconnect)
	local changed = false
	if reconnect then
		host.status = 'Reconnecting'
		forms.settext(formPlayerList, "")
	elseif host.status ~= 'Idle' then
		host.status = 'Idle'
		host.locked = false
		changed = true
	end

	for i,client in pairs(host.clients) do
		client:close()
		host.clients[i] = nil
	end
	host.users = {}
	host.playerlist = {}
	host.client_ping = {}

	if (server ~= nil) then
		server:close()

		local roomstr, err = http.request('https://us-central1-mzm-coop.cloudfunctions.net/destroy' ..
				'?user=' .. config.user ..
				'&pass=' .. config.pass)
		if (err == 200) then
			printOutput('Room closed.')
		else
			printOutput('Error closing room [Code: ' .. (err or '') .. ']')
		end
	end
	server = nil

	if changed then
		if kicked then
			printOutput("You were kicked from the room.")
			kicked = false
		else
			printOutput("Server closed.")
		end
		forms.settext(formPlayerNumber, "")
		if forms.gettext(readyToggle) == "Unready" then
			forms.settext(readyToggle, "Ready")
		end
	end
	forms.settext(formPlayerCount, "...")
	forms.settext(formPlayerList, "")
	forms.settext(formReadyCount, "...")
	forms.setproperty(readyToggle, "Enabled", false)
	updateGUI()

	if hostControls then
		forms.destroy(hostControls)
	end
end


function host.connected()
	for _, _ in pairs(host.clients) do
		return true
	end
	return false
end


function host.reconnecting()
	return host.status == 'Reconnecting'
end


function host.ishost()
	return ((host.connected()) and (host.server ~= nil))
end


--Get the list of Rooms
function host.getRooms()
	local roomstr, err = http.request('https://us-central1-mzm-coop.cloudfunctions.net/getrooms')
	if (err == 200) then
		if (roomstr == '') then
			return false
		else
			return strsplit(roomstr, ',')
		end
	else
		printOutput('Error fetching room list [Code ' .. (err or '') .. ']')
		return false
	end
end

function host.updateHostControls()
	-- update the kick player list
	forms.setdropdownitems(kickPlayerSelect, invert_table(host.users))
end

return host