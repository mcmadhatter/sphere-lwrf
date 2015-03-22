sphere-LWRF
=============

sphere-LWRF is a Ninja Sphere driver for the LightwaveRF products via an LightwaveFR wifilink. It needs the wifilink to transmit the 433Mhz signals.

All the files here are heavily based on the work of Grayda https://github.com/grayda/sphere-orvibo  including his awesome debug script, and the stuff on the orivibo driver.

Feel free to fork, improve or add push requests, this driver is the first time I have used the go language, I doubt it will ever be my language of choice!

Installing
========

Installing from source:
 1. Ensure you're all set for cross-compiling to Linux / ARM and you've run `go get` to get all the necessary packages ([for osx I found this useful][1] )
 2. Follow these links to [enable SSH on your Sphere][2], and [create the necessary folders][3] in the /data folder
 3. Create a sphere-LWRF folder and take ownership by running `mkdir -p /data/sphere/user-autostart/drivers/sphere-LWRF && sudo chown -R ninja.ninja /data/sphere`
 4. If you're on Linux (tested on Ubuntu 14.04 LTS) of OSX (tested on Yosemite), go to the sphere-LWRF src folder and run `bash ./debug.sh`. Select "deploy" from the menu. This bash script will build the binary and copy it to your Sphere.

  [1]: http://stackoverflow.com/questions/12168873/cross-compile-go-on-osx
  [2]: https://developers.ninja/introduction/enable-ssh.html
  [3]: https://developers.ninja/introduction/directory-structure.html


The app should auto-start when your Sphere starts. Everytime it starts up it will get the latest config from the LightwaveRF website, and new devices can then be added by the sphere app in the usual way.  When the app starts it will send a registration packet to the lightwaverf wifilink box, you need to press the <-Yes   button on the wifilink box to register the sphere and allow the sphere to control the lightwaverf stuff.

Bugs / Known Issues
===================


 - This is still an alpha version of the driver if it gets stuck anywhere, green reboot your Sphere

To-Do + Help wanted on
=======================

- Support for more lightwaveRF products
- Support for brightness levels and moods.
- Use the proper sphere config mechanism for LWRF email and pin
- Add support for the energy monitor.
