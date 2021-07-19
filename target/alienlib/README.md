# alienlib

Alienlib implements a low-level library (C ABI compatible) to access some
DrayOS5-specific data from CFW code.

Currently, it contains a few functions that allow user authorization (by
password or pubkey) based on the information stored in the modem configuration.
This is used by our patched version of Dropbear which can then authenticate
connections based on data (password and ssh pubkey) set in the web panel of the
device.

