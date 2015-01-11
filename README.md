sphere-orvibo
=============

sphere-orvibo is a Ninja Sphere driver for the Orvibo S10 / S20 smart sockets (also sold in Australia as Arlec PC180 WiFi sockets, or Bauhn W2 WiFi sockets). These smart sockets can be controlled via your mobile phone and now, your Ninja Sphere!

Building
============

This driver is still in alpha. It works, but barely. Expect it to not work a lot of the time, and not find multiple sockets. It's still in heavy development, so keep that in mind!

This document assumes you have Go already set up and you're ready to cross-compile. It also assumes that you have a Linux machine or a Mac (these instructions haven't been tested on Windows yet, but should work, just make sure to `set` GOOS and GOARCH)

To build from source, go into your workspace folder and run:

`go get github.com/Grayda/sphere-orvibo && go get && GOOS=linux GOARCH=arm go build`

This will retrieve all the necessary libraries, then build the binary and place it in the same folder. When it's built, SSH into your Ninja Sphere and remount the drive as read-write:

`sudo mount -o remount,rw`

Then create a new folder called sphere-orvibo:

`sudo mkdir /opt/ninjablocks/drivers/sphere-orvibo`

Then use `scp` or SFTP to copy the binary into that folder

Binary version
==============

If you don't want to install Go and mess with cross-compilation and such, you can download a pre-compiled binary from here: http://goo.gl/vDsfdO

To install it:

 - Unzip it to a folder on your computer
 - SSH into your Ninja Sphere and run
  - `sudo mount -o remount,rw` to remount the filesystem as read-write
  - `sudo mkdir /opt/ninjablocks/drivers/sphere-orvibo` to create the sphere-orvibo folder for the driver to live in
  - `sudo chmod 755 sphere-orvibo` to make the folder writable
  - If you're on Windows, download pscp from here: http://www.chiark.greenend.org.uk/~sgtatham/putty/download.html . If you're on Linux, ensure `scp` is installed
  - Run this command on your computer: `scp sphere-orvibo ninja@ninjasphere.local:/opt/ninjablocks/drivers/sphere-orvibo` (replace scp with pscp if on Windows). This copies the binary to the correct folder
  - On your sphere, navigate to `/opt/ninjablocks/drivers/sphere-orvibo` and run `./sphere-orvibo --autostart`

Running
=======

There are two ways of running the driver. See https://github.com/ninjasphere/driver-kodi/issues/3#issuecomment-69425372 for both ways, but to get started quickly, run this:

`/opt/ninjablocks/drivers/sphere-orvibo/sphere-orvibo --autostart`

The binary will run and you should see lots of output. In particular, you should look out for "queried":

> !!!T Type: queried

> We've queried. Name is: My Socket

When you see that, open up the Sphere app on your phone and search for new things. You should see your socket appear.

Bugs / Known Issues
===================

 - This driver is still in alpha and may not reliably detect your socket. If it gets stuck anywhere, Ctrl+C out of it and run it again.
 - This driver seems to only find the first socket and then give up. I'm looking into this, pull requests and fix-ups welcomed
 - Needs moar comments. Next version should have more comments
 - Don't use this driver for anything mission critical. If you want to launch nukes or take over the world, be proactive and do it yourself, don't let a $25 WiFi socket do the dirty work. Shame on you.

To-Do
=====

 - Make this driver work with the Orvibo AllOne IR blaster. That might take a bit, as I need to study up on my Ninja Sphere stuff first to work out if this is currently possible
 - FIX ALL THE BUGS!
 - Make this driver more reliable

 Helping out
 ===========

 This is my first ever project in Go. I literally learned the language while porting this code over from node.js. As a result, it's going to be a mess. I hugely appreciate pull requests, so please fork and send in pull requests!

 If you wish to help out in other ways, you could donate hardware and / or beer money. The more Orvibo gear I can purchase, the more neat features I can cram into the driver and the happier you (and I) will be. Please contact me, or donate via Paypal, to grayda [a@t] solid inc [do.t] org
