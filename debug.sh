clear

export AUTHORNAME="Grayda" # If anyone else uses this script, just pop your name here!
export STOPDRIVER="true" # Should we stop the driver on the Sphere? If set to false, saves time when debugging builds
export COPYJSON="true" # Should we copy package.json to the Sphere? Saves a tiny bit of time
export DRIVERNAME="sphere-orvibo" # So I can easily use this script on the next driver!
export DEFAULTHOST="ninjasphere.local" # The default hostname / IP address for the Sphere
export DEFAULTUSERNAME="ninja" # Default username for SSHing into the Sphere
export DEFAULTPASSWORD="temppwd" # Default password for SSHing into the Sphere
export GOPATH=~/sphere-orvibo # Change this to point to your go project.
export GOROOT=~/go # Change this to point to the location of your go installation


function uPrompt { # Prompts the user for vaious information, such as hostname of the Sphere, username and password
  NSIP=$(whiptail --nocancel --backtitle "$AUTHORNAME's $DRIVERNAME helper script" --inputbox "Enter the IP address or hostname of your Ninja Sphere " 0 0 $DEFAULTHOST 3>&1 1>&2 2>&3)
  NSUN=$(whiptail --nocancel --backtitle "$AUTHORNAME's $DRIVERNAME helper script" --inputbox "Enter the username for your Ninja Sphere " 0 0 $DEFAULTUSERNAME 3>&1 1>&2 2>&3)
  NSPW=$(whiptail --nocancel --backtitle "$AUTHORNAME's $DRIVERNAME helper script" --inputbox "Enter the password for your Ninja Sphere " 0 0 $DEFAULTPASSWORD 3>&1 1>&2 2>&3)
}

function saveOptions {
  
}

function dBuild { # Sets env variables and builds the driver
  echo -n "Setting environment variables.. "
  export GOOS=linux
  export GOARCH=arm
  echo "Done!"
  echo -n "Building $DRIVERNAME.. "
  go build
  echo "Done!"
}

function dStop { # Stop the driver on the Sphere if $STOPDRIVER is true
  if [ "$STOPDRIVER" = "true" ] ; then
    echo -n "Stopping $DRIVERNAME on the Sphere.. "
    sshpass -p $NSPW ssh $NSUN@$NSIP "source /etc/profile; NSERVICE_QUIET=true nservice $DRIVERNAME stop"
    echo "Done!"
  else
    echo "Not stopping the driver. If you need this, set STOPDRIVER to true in debug.sh"
  fi
}

function dStart { # Starts the driver on the Sphere
  echo -n "Starting $DRIVERNAME on the Sphere.. "
  sshpass -p $NSPW ssh $NSUN@$NSIP "source /etc/profile; NSERVICE_QUIET=true nservice $DRIVERNAME start"
  echo "Done!"
}

function dCopyBinary { # Copies the binary to the Sphere
  echo -n "Copying binary to Sphere on $NSIP.. "
  sshpass -p $NSPW scp $DRIVERNAME $NSUN@$NSIP:/data/sphere/user-autostart/drivers/$DRIVERNAME/$DRIVERNAME | whiptail --infobox "Copying $DRIVERNAME to Sphere.." 0 0
  echo "Done!"
}

function dCopyJSON { # Copies package.json to the Sphere if $COPYJSON is true
  if [ "$COPYJSON" = "true" ] ; then
    echo -n "Copying package.json to Sphere on $NSIP.. "
    sshpass -p $NSPW scp package.json $NSUN@$NSIP:/data/sphere/user-autostart/drivers/$DRIVERNAME/package.json whiptail --infobox "Copying package.json to Sphere.." 0 0
    echo "Done!"
  else
    echo "Not copying package.json. If you need this, set COPYJSON to true in debug.sh"
  fi
}

# ====================================================================================================================
# ====================================================================================================================

INPUT=$(whiptail --nocancel --backtitle "$AUTHORNAME's $DRIVERNAME helper script" --menu "Please select an option" 0 0 0 \
  "deploy" "Build, copy and run driver on the Sphere"\
  "build" "Build driver only"\
  "debug_build" "Build and copy driver to Sphere, but not run"\
  "copy" "Copy to Sphere only"\
  "restart" "Restart the driver on the Sphere"\
  "test" "Run main.go test for go-orvibo"\
  "exit" "Exit" 3>&1 1>&2 2>&3)

if [ $INPUT = "deploy" ] ; then
  uPrompt # Calls a function that prompts for sphere address, username and password
  dBuild # Sets environment variables and builds the driver using go build
  dStop # Stops the driver if $STOPDRIVER is true
  dCopyBinary # Copies the binary to the Sphere
  dCopyJSON # Copies package.json to the Sphere if $COPYJSON is true
  dStart # Starts the driver again
elif [ $INPUT = "restart" ] ; then
  uPrompt
  dStop
  dStart
elif [ $INPUT = "build" ] ; then
  echo $AUTHORNAME\'s $DRIVERNAME helper script
  echo ------------------------------------
  echo Build Driver
  echo

  dBuild # Sets environment variables and builds the driver using go build
elif [ $INPUT = "test" ] ; then
    echo "Running go-orvibo test.. "
    go run ../go-orvibo/tests/main.go
elif [ $INPUT = "debug_build" ] ; then
  uPrompt # Calls a function that prompts for sphere address, username and password
  dBuild # Sets environment variables and builds the driver using go build
  dStop # Stops the driver if $STOPDRIVER is true
  dCopyBinary # Copies the binary to the Sphere
  dCopyJSON # Copies package.json to the Sphere if $COPYJSON is true
  dCopyJSON
elif [ $INPUT = "debug_run" ] ; then
  echo "Not yet implemented!"
elif [ $INPUT = "copy" ] ; then
  uPrompt # Calls a function that prompts for sphere address, username and password
  dCopyBinary # Copies the binary to the Sphere
  dCopyJSON # Copies package.json to the Sphere if $COPYJSON is true
elif [ $INPUT = "blerg" ] ; then
  INPUT=$(whiptail --nocancel --backtitle "Hello" --menu "Please select an option" 0 0 0 "deploy" "Deploy Driver" "build" "Build driver only" 3>&1 1>&2 2>&3)
   echo "Got $INPUT"
elif [ $INPUT = "exit" ] ; then
 echo
else
  echo No valid command found.
fi
echo
echo -n "Script completed. at "
date +"%I:%M:%S%P"
echo
