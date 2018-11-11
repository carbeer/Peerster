# Peerster
Decentralized Systems Engineering Project

### Run it
To run the project, you need to build the root directory as well as the client directory through `go build`

Afterwards, you can start a Peerster node through `./Peerster <extra parameters>`. I added an extra flag `-runUI` to let a Peerster node serve the GUI on 8080. A client can be run similarly through `./client/client <extra parameters>`. To see a full list of the available extra parameters, use the flag `-h`.

### Current features
- Peers recognize if the peer that they request a file from doesn't have the chunk available (check for empty data). In this case, the repeated tranmission of the DataRequest is stopped.
- If a peer gets a DataRequest but doesn't know the path to the requesting peer, it will send it to the neighboring peer that transmitted the request.  
- Broadcasting excludes the neighboring peer where a message is received from
- Peers check whether a client messages contain any data at all before sending them out