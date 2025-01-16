#!/bin/sh

# Purpose : Uninstaller script

# cd /usr/local/bkup
# ./servicemgr -service stop
# ./servicemgr -service uninstall

# cd ..
# rm -f -r /usr/local/bkup
# rm -f -r /usr/bin/bkupClient

cd $HOME/.backup
sudo ./servicemgr -service stop
sudo env USR="$USER" ./servicemgr -service uninstall

cd ..
rm -f -r $HOME/.backup
sudo rm -f -r /usr/local/bin/bkupClient
