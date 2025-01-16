#!/bin/sh

# Purpose : Installer script

# mkdir /usr/local/bkup
# cp bkupClient /usr/local/bkup
# cp client-background /usr/local/bkup
# cp .env.common /usr/local/bkup
# cp .env.background /usr/local/bkup
# cp servicemgr /usr/local/bkup

# ln -s /usr/local/bkup/bkupClient /usr/bin/

# cd /usr/local/bkup
# echo -e "\nRESTORE_PATH=\"$HOME/Downloads/bkupRestore\"" >> ./.env.common

# ./servicemgr -service install
# ./servicemgr -service start

mkdir $HOME/.backup
cp bkupClient $HOME/.backup
cp client-background $HOME/.backup
cp .env.common $HOME/.backup
cp .env.background $HOME/.backup
cp servicemgr $HOME/.backup

sudo ln -s $HOME/.backup/bkupClient /usr/local/bin/

cd $HOME/.backup
echo -e "\nRESTORE_PATH=\"$HOME/Downloads/bkupRestore\"" >> ./.env.common
export USR="$USER"
sudo env USR="$USER" ./servicemgr -service install
sudo ./servicemgr -service start
echo $USR

