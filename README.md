# Bizhawk Co-op Netplay

bizhawk-co-op is a Lua script for BizHawk that allows two or more people to play a co-op experience by sharing inventory/ammo/hp over the network. This will work with vanilla versions of the games and also randomizers.

## Metroid: Zero Mission Co-op

Health and ammo is shared. Items obtained by a player are given to everyone. Items locations are split such that each item can be optained by only one player. The other players will find a screw attack block in its place. The items at the unknown item locations, power grip location, and imago location can be obtained by either player. Events such as boss deaths are also shared.

## Link to the Past Co-op

Items obtained by a player are given to everyone. Items locations are split such that each item can be optained by only one player. It's usually a good idea to spread out on the overworld and communicate which items still need to be checked by who. With Split Keys or Raid Bosses enabled, it's highly recommended to enter dungeons together. The following can are configurable:
 
* **Health and Ammo** Health and ammo is shared. This includes bottles, bombs, arrows, magic, etc. Death is synced, so if you notice a pause, it's because the script is waiting for everyone to die together
* **Split Big Keys** If enabled then dungeon Big Keys will be split so only one person can obtain it. Otherwise either player can get them.
* **Split Small Keys** Similar to Big Keys, but is for the dungeon Small Keys. This includes Pot Keys and Enemy Drop Keys. Using a key does not remove other players keys. When disabled keys are not shared.
* **Raid Bosses** Bosses share HP. Bosses have scaled HP based on the number players. Players get a damage boost if they attack the boss together at the same time. If everyone is fighting together, then the boss should have similar amount of health as solo. A known bug is that boss that can normally be killed in one hit still die in one hit, so bring your silver arrows with you. You can also have someone stay behind and be the dedicated healer at the cost of dps.

| Players | Boss HP |
|---|-------|
| 1 | 1x    |
| 2 | 3.5x  |
| 3 | 9.5x  |
| 4 | 23.5x |
| 5 | 55.5x |
| ... | ... |

* **Junk Chests** If enabled then item locations not owned by players are filled with random junk items. Having this enabled will make it harder to coordinate who still needs to chest some item location. When disabled they are instead replaced with empty chests. If you find an empty location it means that someone else owns that item location and should check it still.

## Ocarina of Time Cross World Co-op

Nothing is shared, however there are now player specific items which are mixed between all the worlds. So if you obtain an item for yourself only you get it. If you obtain an item for another player then only they get it. This effectively means everyone will be playing different intermingled seeds.

This works with the latest OoT Randomizer found on the website [https://www.ootrandomizer.com](https://www.ootrandomizer.com/) and the latest major release of the source code [GitHub](https://github.com/TestRunnerSRL/OoT-Randomizer/tree/master). Set the Player Count to the number and use the same settings and seed. Each player should then set a unique Player ID (from 1 to the Player Count). The output filename should be the **same** for every player except the last number which indicates the player ID (excluding `-comp`). The logic will guarantee that every player can beat the game.  
* **2-Player File name example:** 
> - `OoT_R4AR3PKKPKF8UK7DSA_TestSeed_W2P1-comp.z64`
> - `OoT_R4AR3PKKPKF8UK7DSA_TestSeed_W2P2-comp.z64` 

## Setup
There are two different methods to install.
* Run the PowerShell script, Download it here: [bizhawk-co-op.ps1](https://github.com/TestRunnerSRL/bizhawk-co-op/releases). I suggest placing it wherever you want to install Bizhawk. To install it, right-click it and select "Run with PowerShell". This will download & install a fresh copy of BizHawk with all the required files in their correct locations.  
**OR**  
* You can manually download the files, install and move them in the correct locations as described below.

### You will need the following:

* (1) [BizHawk 2.6.1](https://github.com/TASVideos/BizHawk/releases/tag/2.6.1)
- The co-op script should be compatible for Bizhawk `1.12.0+` and `2.2.2+`
* (2) [BizHawk prerequisite installer](https://github.com/TASVideos/BizHawk-Prereqs/releases/tag/2.4.8_1) (run this)
* (3) [luasocket](http://files.luaforge.net/releases/luasocket/luasocket/luasocket-2.0.2/luasocket-2.0.2-lua-5.1.2-Win32-vc8.zip)
* (4) [bizhawk-co-op](https://github.com/TestRunnerSRL/bizhawk-co-op/releases)  
> - **If you are running <ins>OoTR 3.0 Release</ins>, replace the file, `bizhawk-co-op/ramcontroller/Ocarina of Time.lua` with this version of the file:** [bizhawk-co-op/ramcontroller/Ocarina of Time.lua](https://github.com/TestRunnerSRL/bizhawk-co-op/raw/9e73e90eae6d5ef82cca96a9f3e6a235abbaf906/bizhawk-co-op/ramcontroller/Ocarina%20of%20Time.lua)  

### Directory structure

The locations of files is very important! Make sure to put them in the right place. After unzipping BizHawk (1), you should be able to find the executable `EmuHawk.exe`, we will call the folder containing it `BizHawkRoot/`.

First, in luasocket (3), you should find three folders, a file, and an executable: `lua/`, `mime/`, `socket/`, `lua5.1.dll`, and `lua5.1.exe`.
Place `mime/` and `socket/` in `BizHawkRoot/`, and place the *contents* of `lua/` in `BizHawkRoot/Lua/`. Place `lua5.1.dll` in `BizHawkRoot/dll/`. You do not need `lua5.1.exe`.

Next, the bizhawk co-op distribution includes two important things: the main lua script `bizhawk co-op.lua` and a folder `bizhawk-co-op/`. Place both of these in `BizHawkRoot/`.

Once this is done, your directory structure should look like this:

```
(1) BizHawk-2.6.1/
(4)   bizhawk-co-op/
(1)   dll/
(3)     lua5.1.dll
        ...
(3)   mime/
        ...
(3)   socket/
        ...
(1)   Lua/
(3)     socket/
(3)     ltn12.lua
(3)     mime.lua
(3)     socket.lua

(4)   bizhawk co-op.lua
(1)   EmuHawk.exe
      ...
```

### bizhawk-co-op Configuration

If using Bizhawk 2.2.2+, go to `Config -> Customize... -> Advanced` and set `Lua Core` to `Lua+LuaInterface`. NLua does not support LuaSockets properly. After changing this setting, you need to close and restart the emulator for the setting to properly update.

Once you have everything else properly set up, you can run the bizhawk-coop script to do some final setup before syncing and playing a game. To run the script in BizHawk, go to `Tools -> Lua Console`, and the Lua Console should open up. At this point, I suggest checking `Settings -> Disable Script on Load` and `Settings -> Autoload`. The former will allow you to choose when to start the script after opening it instead of it running automatically, and the latter will open the Lua Console automatically when you load EmuHawk.

Next, go to `Script -> Open Script...` and open `bizhawk co-op.lua` (it should be in `BizHawk-2.6.1/` root.) Make sure you are running a game, and then double click bizhawk co-op (or click it and then press the green check mark) to run the script. The window has the following important configurations:

* Host IP and Port: The client should set the IP to the host's IP address, and both players must choose the same port number. The <ins>host</ins> will need to enable port forwarding on the chosen port, and will have to make sure their firewall is not blocking BizHawk. As for setting up port forwarding, Google is your best friend. 
> > * <ins>Note:</ins> This may not apply to everyone but make sure you don't have `UPnP IGD` enabled on your router, this setting could prevent you from joining a host or hosting a room. 
> > * <ins>Port forwarding alternative:</ins> "In the event you do not have access to your router to apply port forwarding, try using the program called, "[Hamachi](https://www.vpn.net/)". This program allows you & others to connect to one another as if you are on the same LAN (Local Area Network). Don't let the subscription stuff scare you on their site, all you need is a free account!"

* Game Script: Be sure to choose the appropriate game when creating the room or joining a room.

Make sure to click Save Settings, and you should be ready to play!


## Syncing with bizhawk-coop

The host should first enter their name and password and click `Create Room` to host. Then the clients should click `Refresh` and select the appropriate room, enter their name and the room password, and click `Join Room`. The bizhawk-co-op script will run some consistency checks on your configurations to make sure you are running the same code. If these all passes then the players will be connected. So if you encounter any issues connecting to one another, make sure all players possess an up to date script.

* `Lock Room`: The Host can click this to prevent prevent anyone else from joining the room. 
* `Leave Room`: Click this to cleanly close down the connection. Closing the Lua Console or BizHawk directly can result in issues reconnecting for some time.


### Supported Systems

bizhawk-co-op will only run on a Windows OS because of BizHawk support.

### Credits

Created by TestRunner.

BizHawk, Lua, Luasocket, and kikito's sha1 script. Lua, luasocket, and sha1.lua all fall under the MIT license.

### Issues

If you have any problems with the script (and restarting BizHawk does not fix them,) contact me (TestRunner ([@Test_Runner](https://twitter.com/Test_Runner)) on Twitter or on Discord. You can also submit an issue here on the GitHub.
