sphere-orvibo
=============

sphere-orvibo is a Ninja Sphere driver for the Orvibo S10 / S20 smart sockets (also sold in Australia as Arlec PC180 WiFi sockets, or Bauhn W2 WiFi sockets). These smart sockets can be controlled via your mobile phone and now, your Ninja Sphere!

Installing
==========

Until there's a better way to install drivers on the Sphere, it's a manual process. But for now:

Installing from source:
 1. Ensure you're all set for cross-compiling to Linux / ARM and you've run `go get` to get all the necessary packages
 2. Follow these links to [enable SSH on your Sphere][1], and [create the necessary folders][2] in the /data folder
 3. Create a sphere-orvibo folder and take ownership by running `mkdir -p /data/sphere/user-autostart/drivers/sphere-orvibo && sudo chown -R ninja.ninja /data/sphere`
 4. If you're on Linux (tested on Ubuntu 14.04 LTS), go to the sphere-orvibo src folder and run `bash ./debug.sh`. Select "deploy" from the menu. This bash script will build the binary and copy it to your Sphere.

Installing from a release:
 1. Follow these links to [enable SSH on your Sphere][1].
 2. [Download the latest release][3] from GitHub. You'll need `sphere-orvibo`, `package.json` and `install.sh` (the latter file should work on most Linux distributions, but not Windows or Mac, due to the lack of sshpass and whiptail)
 3. Run `bash ./install.sh`. Follow the instructions. This will create the necessary folders on your Sphere and copy the driver over

Installing release manually
 1. SSH into your Sphere and run `mkdir -p /data/sphere/user-autostart/drivers/sphere-orvibo && sudo chown -R ninja.ninja /data/sphere`
 2. Run `wget https://github.com/Grayda/sphere-orvibo/releases/download/INSERT_VERSION_HERE/sphere-orvibo && wget https://github.com/Grayda/sphere-orvibo/releases/download/INSERT_VERSION_HERE/package.json`. Replace INSERT_VERSION_HERE with the latest version from GitHub (e.g. v0.5.0)
 3. Run `chmod +x /data/sphere/user-autostart/drivers/sphere-orvibo/sphere-orvibo && nservice sphere-orvibo start`to give execute permission to the file and then run it.

  [1]: https://developers.ninja/introduction/enable-ssh.html
  [2]: https://developers.ninja/introduction/directory-structure.html
  [3]: https://github.com/Grayda/sphere-orvibo/releases/latest

Running
=======

The app should auto-start when your Sphere starts. If at any point your sockets stop responding, do a green reset and they should start working again.

Bugs / Known Issues
===================

 - This driver is still in beta and may not reliably detect your socket. If it gets stuck anywhere, green reboot your Sphere
 - Needs moar comments. Next version should have more comments
 - Don't use this driver for anything mission critical. If you want to launch nukes or take over the world, be proactive and do it yourself, don't let a $25 WiFi socket do the dirty work. Shame on you.
 - State may appear stuck on the Android version of the Ninja Sphere. Toggling the socket still shows it as on or off, even when it's not. State is correctly updated in the driver, but not reflected in the app until you exit, or hit refresh on the home screen. This is a known issue with the Ninja Sphere app and will be fixed up later. iOS doesn't have this issue.


To-Do
=======

- Make this driver work with the Orvibo AllOne IR blaster. That might take a bit, as I need to study up on my Ninja Sphere stuff first to work out if this is currently possible
 - FIX ALL THE BUGS!

 Helping out
 ===========

 This is my first ever project in Go. I literally learned the language while porting this code over from node.js. As a result, it's going to be a mess. I hugely appreciate pull requests, so please fork and send in pull requests!

 If you wish to help out in other ways, you could donate hardware and / or beer money. The more Orvibo gear I can purchase, the more neat features I can cram into the driver and the happier you (and I) will be. Please contact me, or donate via Paypal, to grayda [a@t] solid inc [do.t] org
