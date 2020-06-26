remotemoe - everything ssh plumbing

Stuff that needs doing
* service unit with StateDirectory for ACME
* custom hostnames
    * includes the adventure of a database of some kind
* readme
* figure out if ssh.Session.DialContext needs to deal with the provided context
* rewrite of a sessions terminal so that it:
    * isnt just a big ugly switch
    * accepts commands both when a session is active and when a client tries to execute a command directly
    * it should be possible to indicate which names you require in an idempotent way
* proper ssh exit messages

Cool things that should not be done yet
* in the terminal session, have a "debugon" command which provides the user with relevant info about connections being made, http requests etc
* enable users to add other pubkeys which they should be able to manage using any one of the linked keys
* enable users to request random tcp ports for services that cannot mux - for a "1:1 mapping"

Items that need more research:
* instead of buffered ssh.session.msgs - sync .msgs - have the terminal provide it and only let send's happen if non-nil
    * But its properly not as simple as it seems, it would be nice to be able to "buffer" messages which the user will
        receive once (or if) he opens a terminal, given ssh's nature we cannot really know beforehand if a connection is just
        forwards or forwards and a terminal...
    * usecase: we should notice the user if a port is forwarded that we dont know what to do with
