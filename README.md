1. Check command line parameters to determine if in server or CLI mode
2. If Server:
	- Parse in config as a server
    - Parse in wg.conf (filename determined by config at /etc/cayswap)
    - Use key passed in from CLI or from config file for auth
    - Use interface passed in from CLI or from config for listening
    - receive requests from clients
    - on request received, check client IP and check to see if client exists
    - if client exists, error out and write out error message
    - if client is new, then:
    - create new entry in wg config file with IP address + public key
    - systemctl reload wg-quick@
    - afer 15 minutes, exit
3. If Client:
    - Read in key from CLI
    - Read in endpoint from CLI
    - Parse WG Config
    - Post public key + IP to server
    - receive public key + ip back from server
    - update wg config and reload wg-quick, exit
