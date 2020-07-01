package ssh

const firstTimeDiagram = `
      raspberry pi
     +------------------------------------+
     |$ ssh -R22:localhost:22 \           |
     |      -R80:localhost:80 remotemoe   |
     +---------------+^^+-----------------+
                     |**|
 corporate firewall  |**|
+--------------------|**|-----------------------+
                     |**|
                     |**|
                     |**| http and ssh traffic are
                     |**| wrapped inside ssh tunnel
                     |**|
                     |**|
      remotemoe      |**|
     +---------------v--v-----------------+
     |maps services such as http, https   |
     |and ssh to ssh tunnels.             |
     |                                    |
     +-------^----------------------^-----+
             *                      *
             * http traffic         * ssh traffic
 internet    *                      *
+------------*----------------------*-----------+
             *                      *
 browser     *           ssh client *
+------------+---------+ +----------+-----------+
|$ curl key.remotemoe  | |$ ssh -J remotemoe key|
|                      | |                      |
+----------------------+ +----------------------+
`
