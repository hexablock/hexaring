# hexaring [![Build Status](https://travis-ci.org/hexablock/hexaring.svg?branch=master)](https://travis-ci.org/hexablock/hexaring)
Hexaring implements a replicated lookup mechanism on top of chord in order to disassociate
the ring key mapping from the application keys.  It provides various helper functions to
perform ring based operations to allow this.

### Location Based Lookups
Lookups return the specified number of locations for a given key around the ring.  The
algorithm used is as follows:

- Hash a key to compute the natural key
- Get requested number of unique replicas around the ring using the natural key as the
offset
