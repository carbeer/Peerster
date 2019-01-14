#!/usr/bin/env bash

go build
cd client
go build
cd ..

COLOR=$'\e[1;30m'
NC='\033[0m' 
DEBUG="true"

outputFiles=()
message1=Hello!
message2=How_are_you?


file1="file1.txt"
file2="file2.txt"


UIPort=12345
gossipPort=5000
name='A'
rt=5

# General peerster (gossiper) command
#./Peerster -UIPort=12345 -gossipAddr=127.0.0.1:5001 -name=A -peers=127.0.0.1:5002 > A.out &

for i in `seq 1 2`;
do
	outFileName="$name.out"
	peerPort=$((($gossipPort+1)%10+5000))
	peer="127.0.0.1:$peerPort"
	gossipAddr="127.0.0.1:$gossipPort"
	./Peerster -UIPort=$UIPort -gossipAddr=$gossipAddr -peers=$peer -rtimer=$rt> $outFileName &
	outputFiles+=("$outFileName")
	if [[ "$DEBUG" == "true" ]] ; then
		printf "${COLOR}$name running at UIPort $UIPort and gossipPort $gossipPort ${NC}\n"
	fi
	UIPort=$(($UIPort+1))
	gossipPort=$(($gossipPort+1))
	name=$(echo "$name" | tr "A-Y" "B-Z")
done

# Spin up node that serves the UI
./Peerster -UIPort=12355 -gossipAddr=127.0.0.1:5010 -peers=127.0.0.1:5000 -rtimer=$rt -runUI > "UIPeer1.out" &
last_pid=$!

printf "${COLOR}Uploading private file with 4 replications${NC}\n"
./client/client -UIPort=12345 -file=$file1 -private -replications=4

sleep 3

printf "${COLOR}Uploading another private file with 2 replications${NC}\n"
./client/client -UIPort=12346 -file=$file2 -private -replications=2

read varname
printf "${COLOR}Replacing the current UI gossiper with a new instance...${NC}\n"

# Kill the peer that serves the UI
kill -KILL $last_pid
sleep 1

# Spin up new peer that serves the UI
./Peerster -UIPort=12354 -gossipAddr=127.0.0.1:5009 -peers=127.0.0.1:5000 -rtimer=$rt -runUI > "UIPeer2.out" &
sleep 1

read varname
printf "${COLOR}Sending a private message and a private encrypted message to the UI peer...${NC}\n"


# read out name of the UI peer (only needed due to public key generation at runtime)
pubkey=$(awk -F":" '$0~/Generated and/{print $NF;exit;}' ./UIPeer2.out) || true
pubkey="${pubkey#"${pubkey%%[![:space:]]*}"}" || true

# Send private message
./client/client -UIPort=12345 -dest=$pubkey -msg=$message1
./client/client -UIPort=12345 -dest=$pubkey -msg=$message2  -encrypt

read varname
printf "${COLOR}Ending Peerster...${NC}\n "

pkill -f Peerster