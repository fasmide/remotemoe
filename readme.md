```
                            __                              
.----.-----.--------.-----.|  |_.-----.--------.-----.-----.
|   _|  -__|        |  _  ||   _|  -__|        |  _  |  -__|
|__| |_____|__|__|__|_____||____|_____|__|__|__|_____|_____|

remotemoe - ssh plumbing all the things
```

# What is it
remotemoe is a software daemon for exposing ad-hoc services to the internet without having to deal with the regular network stuff such as configuring VPNs, changing firewalls, or adding port forwards. 

Common use-cases include:
* Allow third-party services to access your web app while you're developing it.
* Let containers expose themself to the internet without having to change any infrastructure.
* Quickly share a web app your working on with a collaborator or team for review.
* Allow your CI to run development branches that expose them-self for review.
* Access remotely deployed Raspberry Pi's.

remotemoe doesn't require its users to install, trust, or run any third-party software. It uses plain old SSH, which is available everywhere these days.

# What it's not
It's no SaaS; if you need a reliable service, you're probably going to have to run it your self - any small cloud instance should do just fine...

Available for getting started and testing is `remote.moe`. It is provided with no guarantees and will run broken and unstable branches from time to time :)

# remote.moe - Have a go
Use `remote.moe` if you are ready for a quick and dirty getting started experience. Assume you have a web server running on your local machine that listens for HTTP traffic on port 8080. 

In a terminal, enter: (... or omit if you have something else in mind)
```
$ cd Pictures/; python -m SimpleHTTPServer 8080
Serving HTTP on 0.0.0.0 port 8080 ...
```

In another terminal, enter:
```
$ ssh -R80:localhost:8080 remote.moe
New to remotemoe? - try 'firsttime' or 'help' and start exploring!

http (80)
http://7k3j6g3h67l23j345wennkoc4a2223rhjkba22o77ihzdj3achwa.remote.moe/

$ 
```

That's pretty much all there is to it - all your nudes are now accessible on the URL that remotemoe spits out. 

Next up is typing `help` to have a look at some of the other features. For instance, you could add a more human-friendly hostname, add HTTPS and SSH forwards, or look at the different ways to keep an ssh tunnel open.

# Running your remotemoe
You will need
* Some cloud instance, running ubuntu or similar
* ... that has a public IP address
* ... and a domain or subdomain with records appropriately configured
* Knowledge of Golang and general systems administration :)

To run remotemoe, you need to:

* Fetch this repo, build and move remotemoe to your instance or server
* Create a service for running remotemoe, take inspiration from `infrastructure/remotemoe.service`
* Ensure the hostname of the machine is set accordingly to your domain or subdomain.
* Move openssh out of the way, remotemoe wants to listen on port 22

This will be automated in the future :)

# Compared to Cloudflare's Argo Tunnels
Argo tunnels, and Cloudflare in general, do a lot of things that remotemoe does not, but one similarity is their trycloudflare.com service (https://blog.cloudflare.com/a-free-argo-tunnel-for-your-next-project/) where everyone can expose their web app through a tunnel.

Using their example, when using Argo tunnels, you are required to download their client and run:
```
$ cloudflared tunnel --url localhost:7000
```
a remotemoe equivalent would be:
```
$ ssh -R80:localhost:7000 remote.moe
```

remotemoe and especially Cloudflare does a lot more than this, but to highlight a few differences:

Cloudflare provides a massive Highly Available service at a cost - remotemoe does not.

Cloudflare requires you to create an account if you need to define hostnames or bring a custom domain - remotemoe does not.

remotemoe can be used as an SSH ProxyJump-host and is not limited to any specific protocol - any TCP port is reachable through remotemoe.
