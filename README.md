# Peerster
Decentralized Systems Engineering Project

### Run it
To run the project, you need to build the root directory as well as the client directory through `go build`

Afterwards, you can start a Peerster node through `./Peerster <extra parameters>`. I added an extra flag `-runUI` to let a Peerster node serve the GUI on 8080. A client can be run similarly through `./client/client <extra parameters>`. To see a full list of the available extra parameters, use the flag `-h`.

### Current weaknesses
- The peers must know each other before downloading a file or being able to send a private message. This requires previous rumor messages.

