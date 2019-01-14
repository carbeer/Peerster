# Peerster
## An incentive compatible decentralized file storage
Decentralized Systems Engineering Project.

## Run it
To run the project, you need to build the root directory as well as the client directory through `go build`

Afterwards, you can start a Peerster node through `./Peerster <extra parameters>`. A client can be run similarly through `./client/client <extra parameters>`. Apart from the flags specified in the homeworks, the following flags are available for the Peerster executable:
- `-runUI`: Indicate that the Peerster node is supposed to serve the GUI on port 8080. 

For the client executable:
- `-encrypt`: Use encryption algorithm for the specified file or private message. Attention: For Private Messages, the peer node must have a valid public key as nodeID.
- `-replications=<int>`: Specify the number of replications for a private file upload that shall be distributed among peers in the network
- `-state=<string>`: Indicates that the specified file to be uploaded contains is a state that shall be imported

 To see a full list of the available parameters, use the flag `-h`.
 All functionality can be invoked from the GUI as well.


## Final project implementation
The work done as part of the final project can be mainly found in the following files:
- `asymmEncryption.go`
- `challengeHandler.go` 
- `keyHandling.go` 
- `privateFileDownload.go`
- `privateFileHandler.go`
- `stateManagement.go`
- `symmEncryption.go`

### New user functionality

#### Private Messages 
Can now be sent using RSA content encryption and digital signatures. \
CLI: adding the `-encrypt` flag to a regular private message.

#### Private Files
-  Can be uploaded along with a specification of the number of replications that shall be distributed. \
CLI: adding `-private` and `-replications=<int>` to a regular file upload
- State export (only GUI).
- State upload. \
CLI: Specify `-state=<string>` with the name of the state file
- Download a private file. Automatically determines from where to retrieve the file, i.e. from the local node or from a remote peer. In the ladder case, the file can be found as encrypted version and a decreypted version which is prefixed by `decrypted_` in the download folder. \
CLI: Set the `-encrypt` flag to a regular file download request

### Demo video
The demo video was created and can be reproduced using the `test_final_project.sh` script.