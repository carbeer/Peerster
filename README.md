# Peerster
Decentralized Systems Engineering Project

### Run it
To run the project, you need to build the root directory as well as the client directory through `go build`

Afterwards, you can start a Peerster node through `./Peerster <extra parameters>`. I added an extra flag `-runUI` to let a Peerster node serve the GUI on 8080. A client can be run similarly through `./client/client <extra parameters>`. To see a full list of the available extra parameters, use the flag `-h`.

### Assumptions
- We don't need to store (private) messages sent by a CLI client if they target a destination that we don't know at that point in time.
- All peers start at the same time, as stated in the forum as there is no way of retrospectively requesting blocks. We only accept blockchains as valid, if they eventually link to the genesis block (i.e. Hash of all zeroes). Forks that are detached from the genesis block are not considered valid. 
- We do NOT need to refuse indexing of files that are already part of the blockchain. However, we won't add them as a new transaction to the blockchain (Yes, I am aware that this kind of undermines the idea of avoiding name theft - However, nothing about that was mentioned in the assignment, there was no answer to a forum post (https://moodle.epfl.ch/mod/forum/discuss.php?d=12715) and I asked it in person during an exercise session and got that confirmed, so I hope this is aligned with your expectations. As files are indexed, they are also returned when beind queried through a SearchRequest. )
- We can still accept SearchReplies, even if the search is already finished (by matching the threshold). In my GUI implementation you can see all available files that were found within the network (not only the ones that were found in response to the last SearchRequest).

### Current features
- Peers recognize if the peer that they request a file from doesn't have the chunk available (check for empty data). In this case, the repeated tranmission of the DataRequest is stopped.
- If a peer gets a DataRequest but doesn't know the path to the requesting peer, it will send it to the neighboring peer that transmitted the request.  
- Broadcasting excludes the neighboring peer where a message is received from
- Peers check whether a client messages contain any data at all before sending them out