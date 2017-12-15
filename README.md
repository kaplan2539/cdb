# CHIP debugging bridge (cdb)

The CHIP debugging bridge consists of a daemon (cdb-dameon) running on the embedded device and a command line interface (cdb-cli). The daemon is offering its services via and http API, so the device can be remote-controlled by running the client on any computer. 

## Security
As the cdb-daemon accepts any client connection without authentication, it should not be deployed onto production devices. As a default, the cdb-demon only accepts connections on the two virtual USB-ethernet gadget devices.
