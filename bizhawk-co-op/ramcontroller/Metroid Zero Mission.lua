-- romname: Metroid [-] Zero Mission

-- Mapping for the flags of each ability (little endian)
local AbilityMap = {
	--longbeam = 0
	0x00000001, 	
	--icebeam = 1
	0x00000002,
	--wavebeam = 2
	0x00000004,
	--plasmabeam = 3
	0x00000008,
	--chargebeam = 4
	0x00000010,
	--bombs = 5
	0x00000080,

	--hijump = 6
	0x00010000,
	--speedbooster = 7
	0x00020000,
	--spacejump = 8
	0x00040000,
	--screwattack = 9
	0x00080000,
	--variasuit = 10
	0x00100000,
	--gravitysuit = 11
	0x00200000,
	--morphball = 12
	0x00400000,
	--powergrip = 13
	0x00800000
}

local TankName = {
	[0]="Energy Tank",
	[1]="Missiles",
	[2]="Super Missiles",
	[3]="Power Bombs"
}

local AreaName = {
	[0]="Brinstar",
	"Kraid",
	"Norfair",
	"Ridley",
	"Tourian",
	"Crateria",
	"Chozodia",
	"Area 7"
}

local AbilityName = {
	"Long Beam",
	"Ice Beam",
	"Wave Beam",
	"Plasma Beam",
	"Charge Beam",
	"Bombs",

	"Hi-Jump",
	"Speed Booster",
	"Space Jump",
	"Screw Attack",
	"Varia Suit",
	"Gravity Suit",
	"Morph Ball",
	"Power Grip",
}

local EventName = {
	[0]="Dummy event",
	"Easy mode",
	"Hard mode",
	"Enter Norfair demo played",
	"Exit Kraid demo played",
	"Enter Ridley demo played",
	"Enter Mother Ship demo played",
	"Enter Touriain demo played",
	"Grabbed by Chozo in Brinstar (shows long beam)",
	"Grabbed by Chozo in Brinstar (shows bombs)",
	"Grabbed by Chozo in Brinstar (shows ice beam)",
	"Grabbed by Chozo in Norfair (shows speed booster)",
	"Grabbed by Chozo in Brinstar (shows hi-jump)",
	"Grabbed by Chozo in Norfair (shows Varia suit)",
	"Grabbed by Chozo in Brinstar (shows wave)",
	"Grabbed by Chozo (shows screw attack) [unused]",
	"Power grip obtained",
	"Chozo pillar fully extended",
	"Hi-jump obtained",
	"Varia suit obtained",
	"Charge beam obtained",
	"Screw attack obtained",
	"Space jump obtained",
	"Gravity suit obtained",
	"Plasma beam obtained",
	"Charge beam boss encountered at first location or killed",
	"Charge beam boss encountered at second location or killed",
	"Charge beam boss killed at second location",
	"Acid worm killed",
	"Kraid eyedoor killed",
	"Kraid killed",
	"Kraid elevator statue destroyed",
	"Caterpillar killed",
	"Imago tunnel discovered",
	"Cocoon killed",
	"Imago killed",
	"Ridley eyedoor killed",
	"Ridley killed",
	"Ridley elevator statue destroyed",
	"Mother Brain Killed",
	"Crocomire killed [unused]",
	"Repel machine killed [unused]",
	"Viewed statue room after long beam",
	"Dessgeega killed after statue room",
	"All three hives destroyed",
	"Bugs killed after bombs",
	"Ziplines activated",
	"Plant destroyed (in lava)",
	"Plant destroyed (post-Varia)",
	"Plant destroyed (Varia 2)",
	"Plant destroyed (Varia 3)",
	"Plant destroyed (Varia 1)",
	"Kraid baristutes dead",
	"Kraid statue opened",
	"Ridley statue opened",
	"1st Metroid room cleared",
	"3rd Metroid room cleared",
	"5th Metroid room cleared",
	"2nd Metroid room cleared",
	"6th Metroid room cleared",
	"4th Metroid room cleared",
	"Zebetite 1 destroyed",
	"Zebetite 2 destroyed",
	"Zebetite 3 destroyed",
	"Zebetite 4 destroyed",
	"Escaped Zebes",
	"Marker between Zebes and Mother Ship events",
	"Fully powered suit obtained",
	"Skipped Varia suit",
	"[unknown] (previous implementation of Chozo block?)",
	"Power bomb stolen",
	"Space pirate with power bomb 1",
	"Space pirate with power bomb 2",
	"Glass tube broken",
	"Mecha Ridley killed",
	"Escaped Chozodia",
	"[unknown]",
	"[unknown]",
	"[unused]",
}

itemLocations = {
	{ ID=00, Area=0, Room=0x00, RoomWidth=0x4F, X=0x0b, Y=0x1b, Width=1, Height=1 },
	{ ID=01, Area=0, Room=0x01, RoomWidth=0x13, X=0x0d, Y=0x07, Width=1, Height=1 },
	{ ID=02, Area=0, Room=0x05, RoomWidth=0x13, X=0x07, Y=0x06, Width=2, Height=1 },
	{ ID=03, Area=0, Room=0x02, RoomWidth=0x40, X=0x1c, Y=0x02, Width=1, Height=1 },
	{ ID=04, Area=0, Room=0x29, RoomWidth=0x13, X=0x05, Y=0x12, Width=1, Height=1 },
	{ ID=05, Area=0, Room=0x29, RoomWidth=0x13, X=0x05, Y=0x19, Width=1, Height=1 },
	{ ID=06, Area=0, Room=0x1D, RoomWidth=0x13, X=0x04, Y=0x0a, Width=1, Height=1 },
	{ ID=07, Area=0, Room=0x28, RoomWidth=0x13, X=0x07, Y=0x04, Width=1, Height=1 },
	{ ID=08, Area=0, Room=0x1B, RoomWidth=0x13, X=0x07, Y=0x06, Width=2, Height=1 },
	{ ID=09, Area=-1, Room=-1 },
	{ ID=10, Area=0, Room=0x17, RoomWidth=0x40, X=0x12, Y=0x10, Width=1, Height=1 },
	{ ID=11, Area=0, Room=0x0C, RoomWidth=0x8b, X=0x36, Y=0x06, Width=1, Height=1 },
	{ ID=12, Area=0, Room=0x0F, RoomWidth=0x40, X=0x04, Y=0x06, Width=1, Height=1 },
	{ ID=13, Area=0, Room=0x0E, RoomWidth=0x13, X=0x0e, Y=0x17, Width=1, Height=1 },
	{ ID=14, Area=0, Room=0x13, RoomWidth=0x6d, X=0x0b, Y=0x0a, Width=1, Height=1 },
	{ ID=15, Area=0, Room=0x15, RoomWidth=0x4f, X=0x27, Y=0x05, Width=1, Height=1 },
	{ ID=16, Area=0, Room=0x19, RoomWidth=0x22, X=0x0b, Y=0x05, Width=1, Height=1 },
	{ ID=17, Area=0, Room=0x19, RoomWidth=0x22, X=0x16, Y=0x06, Width=2, Height=1 },
	{ ID=18, Area=0, Room=0x13, RoomWidth=0x6d, X=0x27, Y=0x06, Width=1, Height=1 },
	{ ID=19, Area=1, Room=0x1A, RoomWidth=0x13, X=0x07, Y=0x0a, Width=1, Height=1 },
	{ ID=20, Area=1, Room=0x07, RoomWidth=0x40, X=0x26, Y=0x0e, Width=1, Height=1 },
	{ ID=21, Area=1, Room=0x0A, RoomWidth=0x22, X=0x09, Y=0x09, Width=1, Height=1 },
	{ ID=22, Area=1, Room=0x15, RoomWidth=0x4f, X=0x14, Y=0x03, Width=1, Height=1 },
	{ ID=23, Area=-1, Room=-1 },
	{ ID=24, Area=1, Room=0x08, RoomWidth=0x4f, X=0x4a, Y=0x14, Width=1, Height=1 },
	{ ID=25, Area=1, Room=0x22, RoomWidth=0x13, X=0x06, Y=0x06, Width=2, Height=1 },
	{ ID=26, Area=1, Room=0x26, RoomWidth=0x13, X=0x05, Y=0x04, Width=1, Height=1 },
	{ ID=27, Area=1, Room=0x01, RoomWidth=0x4f, X=0x18, Y=0x04, Width=1, Height=1 },
	{ ID=28, Area=1, Room=0x11, RoomWidth=0x40, X=0x02, Y=0x04, Width=1, Height=1 },
	{ ID=29, Area=1, Room=0x04, RoomWidth=0x22, X=0x16, Y=0x06, Width=1, Height=1 },
	{ ID=30, Area=1, Room=0x09, RoomWidth=0x40, X=0x3c, Y=0x09, Width=1, Height=1 },
	{ ID=31, Area=1, Room=0x02, RoomWidth=0x13, X=0x09, Y=0x21, Width=1, Height=1 },
	{ ID=32, Area=2, Room=0x37, RoomWidth=0x22, X=0x08, Y=0x0e, Width=1, Height=1 },
	{ ID=33, Area=2, Room=0x37, RoomWidth=0x22, X=0x1e, Y=0x17, Width=1, Height=1 },
	{ ID=34, Area=2, Room=0x12, RoomWidth=0x13, X=0x07, Y=0x06, Width=2, Height=1 },
	{ ID=35, Area=2, Room=0x11, RoomWidth=0x22, X=0x11, Y=0x04, Width=1, Height=1 },
	{ ID=36, Area=2, Room=0x01, RoomWidth=0x7c, X=0x41, Y=0x04, Width=1, Height=1 },
	{ ID=37, Area=2, Room=0x03, RoomWidth=0x4f, X=0x48, Y=0x04, Width=1, Height=1 },
	{ ID=38, Area=2, Room=0x1B, RoomWidth=0x13, X=0x07, Y=0x06, Width=2, Height=1 },
	{ ID=39, Area=2, Room=0x26, RoomWidth=0x22, X=0x05, Y=0x06, Width=1, Height=1 },
	{ ID=40, Area=2, Room=0x1C, RoomWidth=0x40, X=0x1c, Y=0x03, Width=1, Height=1 },
	{ ID=41, Area=2, Room=0x1C, RoomWidth=0x40, X=0x36, Y=0x04, Width=1, Height=1 },
	{ ID=42, Area=2, Room=0x25, RoomWidth=0x5e, X=0x15, Y=0x03, Width=1, Height=1 },
	{ ID=43, Area=2, Room=0x20, RoomWidth=0x4f, X=0x04, Y=0x05, Width=1, Height=1 },
	{ ID=44, Area=2, Room=0x20, RoomWidth=0x4f, X=0x2d, Y=0x03, Width=1, Height=1 },
	{ ID=45, Area=2, Room=0x08, RoomWidth=0x22, X=0x15, Y=0x06, Width=2, Height=1 },
	{ ID=46, Area=2, Room=0x0A, RoomWidth=0x31, X=0x0b, Y=0x04, Width=1, Height=1 },
	{ ID=47, Area=2, Room=0x0D, RoomWidth=0x13, X=0x08, Y=0x06, Width=2, Height=1 },
	{ ID=48, Area=2, Room=0x04, RoomWidth=0x5e, X=0x4a, Y=0x09, Width=1, Height=1 },
	{ ID=49, Area=2, Room=0x2F, RoomWidth=0x22, X=0x18, Y=0x03, Width=1, Height=1 },
	{ ID=50, Area=2, Room=0x2A, RoomWidth=0x31, X=0x21, Y=0x05, Width=1, Height=1 },
	{ ID=51, Area=2, Room=0x05, RoomWidth=0x22, X=0x0e, Y=0x4f, Width=1, Height=1 },
	{ ID=52, Area=2, Room=0x05, RoomWidth=0x22, X=0x08, Y=0x6f, Width=1, Height=1 },
	{ ID=53, Area=3, Room=0x1D, RoomWidth=0x22, X=0x18, Y=0x03, Width=1, Height=1 },
	{ ID=54, Area=3, Room=0x1D, RoomWidth=0x22, X=0x14, Y=0x0f, Width=1, Height=1 },
	{ ID=55, Area=3, Room=0x06, RoomWidth=0x13, X=0x08, Y=0x21, Width=1, Height=1 },
	{ ID=56, Area=3, Room=0x0D, RoomWidth=0x22, X=0x08, Y=0x07, Width=1, Height=1 },
	{ ID=57, Area=-1, Room=-1 },
	{ ID=58, Area=3, Room=0x04, RoomWidth=0x13, X=0x06, Y=0x08, Width=1, Height=1 },
	{ ID=59, Area=3, Room=0x17, RoomWidth=0x13, X=0x08, Y=0x04, Width=1, Height=1 },
	{ ID=60, Area=3, Room=0x17, RoomWidth=0x13, X=0x0d, Y=0x0d, Width=1, Height=1 },
	{ ID=61, Area=-1, Room=-1 },
	{ ID=62, Area=3, Room=0x16, RoomWidth=0x13, X=0x0b, Y=0x06, Width=1, Height=1 },
	{ ID=63, Area=3, Room=0x16, RoomWidth=0x13, X=0x08, Y=0x10, Width=1, Height=1 },
	{ ID=64, Area=3, Room=0x12, RoomWidth=0xc7, X=0x48, Y=0x06, Width=1, Height=1 },
	{ ID=65, Area=3, Room=0x09, RoomWidth=0x13, X=0x09, Y=0x04, Width=1, Height=1 },
	{ ID=66, Area=3, Room=0x0A, RoomWidth=0x22, X=0x0f, Y=0x0f, Width=1, Height=1 },
	{ ID=67, Area=3, Room=0x0A, RoomWidth=0x22, X=0x1b, Y=0x06, Width=1, Height=1 },
	{ ID=68, Area=3, Room=0x11, RoomWidth=0x4f, X=0x1c, Y=0x14, Width=1, Height=1 },
	{ ID=69, Area=3, Room=0x0E, RoomWidth=0x40, X=0x1b, Y=0x09, Width=1, Height=1 },
	{ ID=70, Area=3, Room=0x10, RoomWidth=0x40, X=0x36, Y=0x06, Width=1, Height=1 },
	{ ID=71, Area=3, Room=0x1E, RoomWidth=0x22, X=0x04, Y=0x0d, Width=1, Height=1 },
	{ ID=72, Area=3, Room=0x1F, RoomWidth=0x31, X=0x2a, Y=0x07, Width=1, Height=1 },
	{ ID=73, Area=4, Room=0x08, RoomWidth=0x22, X=0x0b, Y=0x6d, Width=1, Height=1 },
	{ ID=74, Area=4, Room=0x07, RoomWidth=0x13, X=0x0e, Y=0x08, Width=1, Height=1 },
	{ ID=75, Area=5, Room=0x05, RoomWidth=0x4f, X=0x14, Y=0x25, Width=1, Height=1 },
	{ ID=76, Area=-1, Room=-1 },
	{ ID=77, Area=5, Room=0x07, RoomWidth=0x5e, X=0x03, Y=0x1b, Width=1, Height=1 },
	{ ID=78, Area=5, Room=0x0E, RoomWidth=0x22, X=0x08, Y=0x0a, Width=1, Height=1 },
	{ ID=79, Area=-1, Room=-1 },
	{ ID=80, Area=5, Room=0x09, RoomWidth=0x5e, X=0x40, Y=0x22, Width=1, Height=1 },
	{ ID=81, Area=5, Room=0x09, RoomWidth=0x5e, X=0x5a, Y=0x09, Width=1, Height=1 },
	{ ID=82, Area=6, Room=0x22, RoomWidth=0x31, X=0x22, Y=0x0e, Width=1, Height=1 },
	{ ID=83, Area=6, Room=0x42, RoomWidth=0x22, X=0x10, Y=0x0d, Width=1, Height=1 },
	{ ID=84, Area=6, Room=0x41, RoomWidth=0x13, X=0x09, Y=0x03, Width=1, Height=1 },
	{ ID=85, Area=6, Room=0x59, RoomWidth=0x13, X=0x06, Y=0x1b, Width=1, Height=1 },
	{ ID=86, Area=6, Room=0x5A, RoomWidth=0x40, X=0x38, Y=0x18, Width=1, Height=1 },
	{ ID=87, Area=6, Room=0x5A, RoomWidth=0x40, X=0x38, Y=0x28, Width=1, Height=1 },
	{ ID=88, Area=6, Room=0x1A, RoomWidth=0x31, X=0x2c, Y=0x08, Width=1, Height=1 },
	{ ID=89, Area=6, Room=0x18, RoomWidth=0x13, X=0x0a, Y=0x0d, Width=1, Height=1 },
	{ ID=90, Area=6, Room=0x36, RoomWidth=0x40, X=0x3b, Y=0x14, Width=1, Height=1 },
	{ ID=91, Area=6, Room=0x2F, RoomWidth=0x13, X=0x09, Y=0x11, Width=1, Height=1 },
	{ ID=92, Area=6, Room=0x31, RoomWidth=0x13, X=0x0a, Y=0x07, Width=1, Height=1 },
	{ ID=93, Area=6, Room=0x5F, RoomWidth=0x22, X=0x18, Y=0x06, Width=1, Height=1 },
	{ ID=94, Area=6, Room=0x4E, RoomWidth=0x31, X=0x2c, Y=0x08, Width=1, Height=1 },
	{ ID=95, Area=6, Room=0x49, RoomWidth=0x13, X=0x09, Y=0x06, Width=1, Height=1 },
	{ ID=96, Area=6, Room=0x0E, RoomWidth=0x13, X=0x0d, Y=0x05, Width=1, Height=1 },
	{ ID=97, Area=6, Room=0x0A, RoomWidth=0x22, X=0x13, Y=0x04, Width=1, Height=1 },
	{ ID=98, Area=6, Room=0x47, RoomWidth=0x40, X=0x3b, Y=0x13, Width=1, Height=1 },
	{ ID=99, Area=6, Room=0x57, RoomWidth=0x31, X=0x12, Y=0x12, Width=1, Height=1 },
}









-- Writes value to RAM using little endian
local prevDomain = ""
function writeRAM(domain, address, size, value)
	-- update domain
	if (prevDomain ~= domain) then
		prevDomain = domain
		if not memory.usememorydomain(domain) then
			return
		end
	end

	-- default size short
	if (size == nil) then
		size = 2
	end

	if (value == nil) then
		return
	end

	if size == 1 then
		memory.writebyte(address, value)
	elseif size == 2 then
		memory.write_u16_le(address, value)
	elseif size == 4 then
		memory.write_u32_le(address, value)
	end
end

-- Reads a value from RAM using little endian
function readRAM(domain, address, size)
	-- update domain
	if (prevDomain ~= domain) then
		prevDomain = domain
		if not memory.usememorydomain(domain) then
			return
		end
	end

	-- default size short
	if (size == nil) then
		size = 2
	end

	if size == 1 then
		return memory.readbyte(address)
	elseif size == 2 then
		return memory.read_u16_le(address)
	elseif size == 4 then
		return memory.read_u32_le(address)
	end
end



-- Gets the list of the abilities in RAM
function getAbility()
	local ability = {}
	flags = readRAM("IWRAM", 0x153C, 4)

	for abilityNum, value in pairs(AbilityMap) do
		-- if ability flags found add ability
		if (bit.band(flags, value) > 0) then
			value = bit.lshift(value, 8)
			active = (bit.band(flags, value) > 0)
			ability[abilityNum] = active
		end
	end
	return ability
end

-- Gets the list of all the event states
function getEvents()
	local events = {}
	for i=0,11 do
		local eventbyte = readRAM("EWRAM", 0x037E00 + i, 1)
		for j=0,7 do
			-- set event number true if bit is set, otherwise false
			events[j + i*8] = (bit.band(eventbyte, 0x01) > 0)
			eventbyte = bit.rshift(eventbyte, 1)
		end
	end
	return events
end

-- Gets the list of ammo values and capacities
function getAmmo()
	return {
		energyCapacity = readRAM("IWRAM", 0x1530, 2),
		missileCapacity = readRAM("IWRAM", 0x1532, 2),
		superCapacity = readRAM("IWRAM", 0x1534, 1),
		powerCapacity = readRAM("IWRAM", 0x1535, 1),

		energyCount = readRAM("IWRAM", 0x1536, 2),
		missileCount = readRAM("IWRAM", 0x1538, 2),
		superCount = readRAM("IWRAM", 0x153A, 1),
		powerCount = readRAM("IWRAM", 0x153B, 1)
	}
end


-- Event to check if a new tank is collected
-- Reverts changes if the tank has been collected already
-- Does not send ammo updates if new tank is found
local prevCollectingTankFlag = 0
function eventTankCollected(prevRam, newRam)
	local collectingTankFlag = readRAM("IWRAM", 0x0044, 1)
	if prevCollectingTankFlag ~= collectingTankFlag then
		prevCollectingTankFlag = collectingTankFlag
		if collectingTankFlag == 1 then
			-- If a tank is collected
			local newTank = {
				tankType = readRAM("IWRAM", 0x56A8, 1) - 0x38,
				areaID = readRAM("IWRAM", 0x0054, 1),
				roomID = readRAM("IWRAM", 0x0055, 1),
				tankX = readRAM("IWRAM", 0x56AA, 1),
				tankY = readRAM("IWRAM", 0x56AB, 1)
			}

			-- Check if it's a new tank
			for i, tank in pairs(prevRam.tanks) do
				local match = true
				for k,v in pairs(tank) do
					if (newTank[k] ~= v) then
						match = false
						break
					end
				end

				if match then
					-- If tank was alread collected then revert changes
					for ammo,value in pairs(prevRam.ammo) do
						newRam.ammo[ammo] = value
					end

					-- Update RAM to previous values
					writeRAM("IWRAM", 0x1530, 2, newRam.ammo.energyCapacity)
					writeRAM("IWRAM", 0x1532, 2, newRam.ammo.missileCapacity)
					writeRAM("IWRAM", 0x1534, 1, newRam.ammo.superCapacity)
					writeRAM("IWRAM", 0x1535, 1, newRam.ammo.powerCapacity)
					writeRAM("IWRAM", 0x1536, 2, newRam.ammo.energyCount)
					writeRAM("IWRAM", 0x1538, 2, newRam.ammo.missileCount)
					writeRAM("IWRAM", 0x153A, 1, newRam.ammo.superCount)
					writeRAM("IWRAM", 0x153B, 1, newRam.ammo.powerCount)

					return false, prevRam, newRam
				end
			end

			-- Add tank to list of tanks
			table.insert(newRam.tanks, newTank)
			return newTank, prevRam, newRam
		end
	end
	-- No tank was collected
	return false, prevRam, newRam
end

-- Event to check when a new ability is collected
function eventAbilityCollected(prevRam, newRam)
	-- Find changed ability
	-- Only one ability can be collected at a time
	-- Only checks for added abilities, not removed (varia)
	local abilityNum = -1
	local abilityActive = true
	for ability,on in pairs(newRam.ability) do
		if (prevRam.ability[ability] ~= on) then
			abilityNum = ability
			abilityActive = on
			break
		end
	end
	
	if abilityNum ~= -1 then
		-- Return the new ability
		return {
			ability = abilityNum,
			active = abilityActive,
			areaID = readRAM("IWRAM", 0x0054, 1),
			roomID = readRAM("IWRAM", 0x0055, 1),
			minimapX = readRAM("IWRAM", 0x0059, 1),
			minimapY = readRAM("IWRAM", 0x005A, 1)
		}
	end
	-- No new ability
	return false
end

-- Event to check if any game events have changed
function eventTriggerEvent(prevRam, newRam)
	local events = {}
	local changed = false

	-- check if any changes
	for i=0,95 do
		if (prevRam.events[i] ~= newRam.events[i]) then
			if newRam.events[i] then
				events[i] = true
			else
				events[i] = false
			end
			changed = true
		end
	end

	if changed then
		-- return list of changes
		return events
	else 
		-- no events changed
		return false
	end
end

-- Event to check if any ammo changed
-- Will not have any changes if a tank was collected
function eventAmmoChange(prevRam, newRam)
	local deltaammo = {}
	local changed = false

	if (prevRam.ammo.energyCount > 0) then
		-- If alive, send delta changes. Don't check for capacities (handled in tank event)
		deltaammo.delta = true

		-- Check energy capacity changes
		if (newRam.ammo.energyCapacity ~= prevRam.ammo.energyCapacity) then
			deltaammo.energyCapacity = newRam.ammo.energyCapacity - prevRam.ammo.energyCapacity
			changed = true			
		end
		-- Check missile capacity changes
		if (newRam.ammo.missileCapacity ~= prevRam.ammo.missileCapacity) then
			deltaammo.missileCapacity = newRam.ammo.missileCapacity - prevRam.ammo.missileCapacity
			changed = true			
		end
		-- Check super capacity changes
		if (newRam.ammo.superCapacity ~= prevRam.ammo.superCapacity) then
			deltaammo.superCapacity = newRam.ammo.superCapacity - prevRam.ammo.superCapacity
			changed = true			
		end
		-- Check power capacity changes
		if (newRam.ammo.powerCapacity ~= prevRam.ammo.powerCapacity) then
			deltaammo.powerCapacity = newRam.ammo.powerCapacity - prevRam.ammo.powerCapacity
			changed = true			
		end		

		-- Check energy count changes
		if (newRam.ammo.energyCount ~= prevRam.ammo.energyCount) then
			deltaammo.energyCount = newRam.ammo.energyCount - prevRam.ammo.energyCount
			changed = true			
		end
		-- Check missile count changes
		if (newRam.ammo.missileCount ~= prevRam.ammo.missileCount) then
			deltaammo.missileCount = newRam.ammo.missileCount - prevRam.ammo.missileCount
			changed = true			
		end
		-- Check super missile count changes
		if (newRam.ammo.superCount ~= prevRam.ammo.superCount) then
			deltaammo.superCount = newRam.ammo.superCount - prevRam.ammo.superCount
			changed = true			
		end
		-- Check power bomb count changes
		if (newRam.ammo.powerCount ~= prevRam.ammo.powerCount) then
			deltaammo.powerCount = newRam.ammo.powerCount - prevRam.ammo.powerCount
			changed = true			
		end
	else 
		-- Was dead, send override values. Check counts AND capacities
		deltaammo.delta = false
		for ammo, value in pairs(newRam.ammo) do
			if (prevRam.ammo[ammo] ~= value) then
				deltaammo[ammo] = value
				changed = true			
			end
		end		
	end

	if changed then
		-- return any changes
		return deltaammo
	else 
		-- ammo is unchanged
		return false
	end
end


-- This sets a tank to be collected and give the appropriate ammo
-- Does not trigger for tanks that have already been collected
function setTankCollected(prevRAM, newTank)
	-- Check if tank was already collected
	for i, tank in pairs(prevRAM.tanks) do
		local match = true
		for k,v in pairs(tank) do
			if (newTank[k] ~= v) then
				match = false
				break
			end
		end

		if match then
			-- If old tank, then do nothing
			return prevRAM
		end
	end

	-- New tank found
	table.insert(prevRAM.tanks, newTank)

	-- Add tank to collected tanks in RAM
	local addr = 0x036C00 + newTank.areaID * 0x100
	while readRAM("EWRAM", addr, 1) ~= 0xFF do
		addr = addr + 4
	end
	
	writeRAM("EWRAM", addr + 0, 1, newTank.roomID)
	writeRAM("EWRAM", addr + 1, 1, newTank.tankType)
	writeRAM("EWRAM", addr + 2, 1, newTank.tankX)
	writeRAM("EWRAM", addr + 3, 1, newTank.tankY)

	writeRAM("IWRAM", 0x0063 + newTank.areaID, 1,
		readRAM("IWRAM", 0x0063 + newTank.areaID, 1) + 1)
	
	-- TODO: update minimap
	-- TODO: remove tank if in same room

	-- Return changes
	return prevRAM
end

-- Set an ability to be collected
function setAbilityCollected(prevAbility, newAbility)
	-- Set ability collected/equipped
	if (newAbility.active) then
		newabilityflag = bit.bor(AbilityMap[newAbility.ability],
			bit.lshift(AbilityMap[newAbility.ability], 8))
	else
		newabilityflag = AbilityMap[newAbility.ability]
	end

	writeRAM("IWRAM", 0x153C, 4, 
		bit.bor(readRAM("IWRAM", 0x153C, 4), newabilityflag))

	prevAbility[newAbility.ability] = newAbility.active

	-- TODO: update VRAM for new beam
	-- if (abilityNum <= 2) then
	--
	-- do

	-- TODO: update minimap
	-- TODO: remove item if in same room

	return prevAbility
end

function removeTankFromRoom()
	-- TODO
end

function removeAbilityFromRoom()
	-- TODO
end

-- Set a game event state to new state
function setEvent(prevEvent, newEvent)
	-- for each even change...
	for event,active in pairs(newEvent) do
		local address = 0x037E00 + math.floor(event / 8)
		local eventflag = bit.lshift(0x01, event % 8)

		local eventbyte = readRAM("EWRAM", address, 1)
		if active then
			-- Set event to active
			eventbyte = bit.bor(eventbyte, eventflag)
		else
			-- Set event to inactive
			eventbyte = bit.band(eventbyte, bit.bnot(eventflag))
		end

		-- Update RAM
		prevEvent[event] = active
		writeRAM("EWRAM", address, 1, eventbyte)
	end

	return prevEvent
end

-- Set ammo counts to new updates
function setAmmo(prevAmmo, deltaAmmo)
	local newAmmo = {}

	if deltaAmmo.delta then
		-- If incremental delta changes, add values to current values
		-- deltas may be negative to subtract
		-- bound the updated values with the capacity and 0
		newAmmo.energyCapacity = math.max(prevAmmo.energyCapacity + 
			(deltaAmmo.energyCapacity or 0), 0)
		newAmmo.missileCapacity = math.max(prevAmmo.missileCapacity + 
			(deltaAmmo.missileCapacity or 0), 0)
		newAmmo.superCapacity = math.max(prevAmmo.superCapacity + 
			(deltaAmmo.superCapacity or 0), 0)
		newAmmo.powerCapacity = math.max(prevAmmo.powerCapacity + 
			(deltaAmmo.powerCapacity or 0), 0)

		newAmmo.energyCount = math.max(math.min(prevAmmo.energyCount + 
			(deltaAmmo.energyCount or 0), newAmmo.energyCapacity), 0)
		newAmmo.missileCount = math.max(math.min(prevAmmo.missileCount + 
			(deltaAmmo.missileCount or 0), newAmmo.missileCapacity), 0)
		newAmmo.superCount = math.max(math.min(prevAmmo.superCount + 
			(deltaAmmo.superCount or 0), newAmmo.superCapacity), 0)
		newAmmo.powerCount = math.max(math.min(prevAmmo.powerCount + 
			(deltaAmmo.powerCount or 0), newAmmo.powerCapacity), 0)
	else
		-- If override changes, set the new value discarding the old value
		for ammo,value in pairs(prevAmmo) do
			newAmmo[ammo] = deltaAmmo[ammo] or value
		end

	end

	-- Update the counts in RAM
	writeRAM("IWRAM", 0x1530, 2, newAmmo.energyCapacity)
	writeRAM("IWRAM", 0x1532, 2, newAmmo.missileCapacity)
	writeRAM("IWRAM", 0x1534, 1, newAmmo.superCapacity)
	writeRAM("IWRAM", 0x1535, 1, newAmmo.powerCapacity)
	writeRAM("IWRAM", 0x1536, 2, newAmmo.energyCount)
	writeRAM("IWRAM", 0x1538, 2, newAmmo.missileCount)
	writeRAM("IWRAM", 0x153A, 1, newAmmo.superCount)
	writeRAM("IWRAM", 0x153B, 1, newAmmo.powerCount)

	return newAmmo
end

-- Object that exposes the public functions
local mzm_ram = {}

-- RAM state from previous frame
local prevRAM = {
	ammo = {
		energyCount = 0,
		energyCapacity = 0,
		missileCount = 0,
		missileCapacity = 0,
		superCount = 0,
		superCapacity = 0,
		powerCount = 0,
		powerCapacity = 0
	},

	tanks = {},
	ability = {},
	events = {}
}



local splitItems = {}
function removeItems()
	areaID = readRAM("IWRAM", 0x0054, 1)
	roomID = readRAM("IWRAM", 0x0055, 1)

	for _,itemLocation in pairs(itemLocations) do
		if splitItems[itemLocation.ID] ~= my_ID and
			(itemLocation.Area == areaID and itemLocation.Room == roomID) then
			for x = itemLocation.X, (itemLocation.X + itemLocation.Width - 1) do
				for y = itemLocation.Y, (itemLocation.Y + itemLocation.Height - 1) do
					-- BG1 Data
					writeRAM("EWRAM", ((x + (y * itemLocation.RoomWidth)) * 2) + 0x2D800, 1, 0x45)
					-- Clip Data
					writeRAM("EWRAM", ((x + (y * itemLocation.RoomWidth)) * 2) + 0x27800, 1, 0x10)
				end
			end
		end
	end
end


-- Gets a message to send to the other player of new changes
-- Returns the message as a dictionary object
-- Returns false if no message is to be send
function mzm_ram.getMessage()
	removeItems()

	-- Gets the current RAM state
	local newRAM = {
		tanks = prevRAM.tanks,
		ability = getAbility(),
		events = getEvents(),
		ammo = getAmmo()
	}

	local message = {}
	local changed = false
	local newTank

	-- Gets the message for a new collected tank
	-- Also updates the states to squelch some changes
	newTank, prevRAM, newRAM = eventTankCollected(prevRAM, newRAM)
	if newTank then
		-- Add new changes
		message["t"] = newTank
		changed = true

		gui.addmessage(config.user .. ": " .. TankName[newTank.tankType] .. " - " .. 
			AreaName[newTank.areaID] .. " [" .. newTank.tankX .. "," .. newTank.tankY .. "]")
	end

	-- Gets the message for a new collected ability
	local newAbility = eventAbilityCollected(prevRAM, newRAM)
	if newAbility then
		-- Add new changes
		message["a"] = newAbility
		changed = true

		gui.addmessage(config.user .. ": " .. AbilityName[newAbility.ability] .. " " .. (message["a"].active and "On" or "Off") .. " - " .. 
			AreaName[newAbility.areaID] .. " [" .. newAbility.minimapX .. "," .. newAbility.minimapY .. "]")
	end

	-- Gets the message for all changed game events
	local newEvent = eventTriggerEvent(prevRAM, newRAM)
	if newEvent then
		-- Add new changes
		message["e"] = newEvent
		changed = true

		for event,active in pairs(newEvent) do
			if (EventName[event] == nil) then
				gui.addmessage(config.user .. ": Event #" .. event .. " - " .. (active and "On" or "Off"))
			else
				gui.addmessage(config.user .. ": Event " .. EventName[event] .. " - " .. (active and "On" or "Off"))
			end
		end
	end

	-- Gets the message for all updated ammo count/capacity
	local newAmmo = eventAmmoChange(prevRAM, newRAM)
	if newAmmo then
		-- Add new changes
		message["m"] = newAmmo
		changed = true
	end

	-- Update the frame pointer
	prevRAM = newRAM

	if changed then
		-- Send message
		return message
	else 
		-- No updates, no message
		return false
	end
end

-- Process a message from another player and update RAM
function mzm_ram.processMessage(their_user, message)
	-- Process new tank collected
	-- Does nothing if tank was already collected
	if message["t"] then
		prevRAM = setTankCollected(prevRAM, message["t"])

		gui.addmessage(their_user .. ": " .. TankName[message["t"].tankType] .. " - " .. 
			AreaName[message["t"].areaID] .. " [" .. message["t"].tankX .. "," .. message["t"].tankY .. "]")
	end

	-- Process new ability collected
	if message["a"] then
		prevRAM.ability = setAbilityCollected(prevRAM.ability, message["a"])

		gui.addmessage(their_user .. ": " .. AbilityName[message["a"].ability] .. " " .. (message["a"].active and "On" or "Off") .. " - " .. 
			AreaName[message["a"].areaID] .. " [" .. message["a"].minimapX .. "," .. message["a"].minimapY .. "]")
	end

	-- process all changed game events
	if message["e"] then
		prevRAM.events = setEvent(prevRAM.events, message["e"])

		for event,active in pairs(message["e"]) do
			if (EventName[event] == nil) then
				gui.addmessage(their_user .. ": Event #" .. event .. " - " .. (active and "On" or "Off"))
			else
				gui.addmessage(their_user .. ": Event " .. EventName[event] .. " - " .. (active and "On" or "Off"))
			end
		end
	end

	-- process all ammo updates
	if message["m"] then
		prevRAM.ammo = setAmmo(prevRAM.ammo, message["m"])
	end

	if message["i"] then
		splitItems = message["i"]
	end
end

mzm_ram.itemcount = 100

return mzm_ram