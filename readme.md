```
                            __                              
.----.-----.--------.-----.|  |_.-----.--------.-----.-----.
|   _|  -__|        |  _  ||   _|  -__|        |  _  |  -__|
|__| |_____|__|__|__|_____||____|_____|__|__|__|_____|_____|

remotemoe - ssh plumber
```


Stuff that needs doing
* custom hostnames
    * includes the adventure of a database of some kind
    * timeouts
    * delete 
    * garbage collection
    * Maybe the timeout should be access based? ... gold would properly be both access'es and active connections
* readme
* rewrite of a sessions terminal so that it:
    * isnt just a big ugly switch
    * accepts commands both when a session is active and when a client tries to execute a command directly
    * it should be possible to indicate which names you require in an idempotent way
* proper ssh exit messages
* tab complete
* maybe dont allow acme to create certificate requests for hosts that do not provide https
* ssh.Terminal provides a TabCompletionCallback which we should use


Cool things that should not be done yet
* in the terminal session, have a "debugon" command which provides the user with relevant info about connections being made, http requests etc
* enable users to add other pubkeys which they should be able to manage using any one of the linked keys
* enable users to request random tcp ports for services that cannot mux - for a "1:1 mapping"

Items that need more research:

* current namedRoute design allows users to steal each others raw pubkey fqdn's 
    * maybe dont allow exactly the length of pubkey.hostname
        * i like this one because they are soo long that no one will acturlly create such a long named route - unless trying todo what we dont want them to
            * and if they do, they are just going to have to add or remove another char...
    * maybe introduce a pattern where the router ensures no-one makes names on a special subdomain
        * base32.k.remote.moe
        * kind of complex
    * maybe dont care: we will check if a route exists when users setup names and we could ensure that router.Replace removes any existing routes .. which is sorta already does but that really only had effect on other connections with the same pubkey
    * maybe check to see if the name that is requested is a keyname we have seen before - it would become quite a long list though

* figure out if ssh.Session.DialContext needs to deal with the provided context

* instead of buffered ssh.session.msgs - sync .msgs - have the terminal provide it and only let send's happen if non-nil
    * But its properly not as simple as it seems, it would be nice to be able to "buffer" messages which the user will
        receive once (or if) he opens a terminal, given ssh's nature we cannot really know beforehand if a connection is just
        forwards or forwards and a terminal...
    * usecase: we should notice the user if a port is forwarded that we dont know what to do with
