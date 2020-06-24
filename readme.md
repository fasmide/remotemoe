remotemoe

Stuff that needs doing
* figure out a place to store the hostname where code can look it up
* remotemoe as jumphost
* HTTPS
* custom hostnames
    * includes the adventure of a database of some kind
* readme
* instead of buffered ssh.session.msgs - sync .msgs - have the terminal provide it and only let other send happen if non-nil
* figure out if ssh.Session.DialContext needs to deal with the provided context
* rewrite of a sessions terminal so that it:
    * isnt just a big ugly switch
    * accepts commands both when a session is active and when a client tries to execute a command directly

Cool things that should not be done yet
* in the terminal session, have a "debugon" command which provides the user with relevant info about connections being made, http requests etc
* enable users to add other pubkeys which they should be able to manage using any one of the linked keys