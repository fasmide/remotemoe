```
                            __                              
.----.-----.--------.-----.|  |_.-----.--------.-----.-----.
|   _|  -__|        |  _  ||   _|  -__|        |  _  |  -__|
|__| |_____|__|__|__|_____||____|_____|__|__|__|_____|_____|

remotemoe - ssh plumber
```


Stuff that needs doing
* readme
* accept commands both when a session is active and when a client tries to execute a command directly
* proper ssh exit messages
* tab complete
* maybe dont allow acme to create certificate requests for hosts that do not provide https
* ssh.Terminal provides a TabCompletionCallback which we should use
* a windows way of keeping the tunnel open
* maybe even a macos way of keeping the tunnel open


Cool things that should not be done yet
* in the terminal session, have a "debugon" command which provides the user with relevant info about connections being made, http requests etc
* enable users to add other pubkeys which they should be able to manage using any one of the linked keys
* enable users to request random tcp ports for services that cannot mux - for a "1:1 mapping"
* clear the database of hostnames that have not been used for a long time

Items that need more research:

* current router design leaves readers waiting when someone is editing the "routing table" - could we have a design where
    a sort of "next routing table" is maintained and it is the only table that is made changes to - then when
    a "update" is available we replace this new table with the old one?

* figure out if ssh.Session.DialContext needs to deal with the provided context

* instead of buffered ssh.session.msgs - sync .msgs - have the terminal provide it and only let send's happen if non-nil
    * But its properly not as simple as it seems, it would be nice to be able to "buffer" messages which the user will
        receive once (or if) he opens a terminal, given ssh's nature we cannot really know beforehand if a connection is just
        forwards or forwards and a terminal...
    * usecase: we should notice the user if a port is forwarded that we dont know what to do with
