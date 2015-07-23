driver-orvibo
=============

driver-orvibo (formerly known as sphere-orvibo) is a Ninja Sphere driver for the Orvibo S10 / S20 smart sockets, plus the AllOne IR blaster. These devices can be controlled via your mobile phone and now, your Ninja Sphere!

Installing
==========

To install this driver from a release:

 1. [Connect your Sphere via USB, and / or enable SSH][1]
 2. SSH in to your Sphere and run the following command:

`sudo with-rw bash`
`wget -qO- https://github.com/Grayda/sphere-orvibo/releases/download/v0.6.0/web-install.sh | bash`

This will download an install script from GitHub and take care of removing old versions, creating folders, downloading the driver and getting it started


  [1]: https://developers.ninja/introduction/enable-ssh.html
  [2]: https://developers.ninja/introduction/directory-structure.html


Running
=======


The app should auto-start when your Sphere starts. If at any point your sockets stop responding, do a green reset and they should start working again. To program and play back IR codes, visit the Labs page in the iOS app, or http://ninjasphere.local in any browser. If an AllOne is detected, you'll see an option to configure IR codes.

Bugs / Known Issues
===================

 - This driver is still in beta and may not reliably detect your socket. If it gets stuck anywhere, green reboot your Sphere
 - Needs moar comments. Next version should have more comments
 - Don't use this driver for anything mission critical. If you want to launch nukes or take over the world, be proactive and do it yourself, don't let a $25 WiFi socket do the dirty work. Shame on you.
 - State may appear stuck on the Android version of the Ninja Sphere. Toggling the socket still shows it as on or off, even when it's not. State is correctly updated in the driver, but not reflected in the app until you exit, or hit refresh on the home screen. This is a known issue with the Ninja Sphere app and is unlikely to be fixed. iOS doesn't have this issue.
 - This version doesn't do 433mhz codes (i.e. the RF switches). I don't have the hardware to add this feature in.


To-Do
=======

 - FIX ALL THE BUGS!
 - Get 433mhz support working, for possible compatibility with the Ninja Blocks
 - Add Kepler support

Helping out
===========

This is my first ever project in Go. I literally learned the language while porting this code over from node.js. As a result, there may be coding errors or bugs present. Pull requests and forks are very much appreciated.

If you're not a coder, you can still help out. I'm accepting donations of hardware and / or beer money. There are plenty of Orvibo products out there, including the Kepler gas detector, the Orvibo RF switch (which works with the AllOne) and some new products on the horizon, like a powerstrip version of the sockets. These cost money, so if you own these and don't want them any more, please contact me. Any money donated goes directly towards the development of this driver, including hardware purchase, installation (for example, the RF switch requires installation by a certified electrician) and any out-of-pocket expenses directly related to the project. Please contact me using my GitHub email address to work out how you can help out.
