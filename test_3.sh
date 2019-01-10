#!/usr/bin/env bash

go build
cd client
go build
cd ..

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m'
DEBUG="true"

outputFiles=()
message_c1_1=Weather_is_clear
message_c2_1=Winter_is_coming
message_c1_2=No_clouds_really
message_c2_2=Let\'s_go_skiing
message_c3=Is_anybody_here?


UIPort=12345
gossipPort=5000
name='A'

# General peerster (gossiper) command
#./Peerster -UIPort=12345 -gossipAddr=127.0.0.1:5001 -name=A -peers=127.0.0.1:5002 > A.out &

for i in `seq 1 3`;
do
	outFileName="$name.out"
	peerPort=$((($gossipPort+1)%10+5000))
	peer="127.0.0.1:$peerPort"
	gossipAddr="127.0.0.1:$gossipPort"
	./Peerster -UIPort=$UIPort -gossipAddr=$gossipAddr -peers=$peer > $outFileName &
	outputFiles+=("$outFileName")
	if [[ "$DEBUG" == "true" ]] ; then
		echo "$name running at UIPort $UIPort and gossipPort $gossipPort"
	fi
	UIPort=$(($UIPort+1))
	gossipPort=$(($gossipPort+1))
	name=$(echo "$name" | tr "A-Y" "B-Z")
done

echo "Finished initial spinup"
./Peerster -UIPort=12355 -gossipAddr=127.0.0.1:5010 -peers=127.0.0.1:5000 -runUI > "UIPeer1.out" &
last_pid=$!

echo "Sending client messages"
./client/client -UIPort=12345 -msg=$message_c1_1
./client/client -UIPort=12346 -msg=$message_c2_1
./client/client -UIPort=12345 -file="hjbfkjbf.txt" -private -replications=4

sleep 5

./client/client -UIPort=12345 -file="bla.txt" -private -replications=3
./client/client -UIPort=12346 -file="lol.txt" -private -replications=2

echo "Press any key to kill the current UI gossiper..."
read varname

kill -KILL $last_pid
sleep 3
./Peerster -UIPort=12354 -gossipAddr=127.0.0.1:5009 -peers=127.0.0.1:5000 -runUI > "UIPeer2.out" &

echo "Press any key to kill the processes..."
read varname

pkill -f Peerster