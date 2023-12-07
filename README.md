# This is a simple key-value in-memory database which supports communication over http
NOTE: it was created in recreational purposes and for no reason is safe to use 
(passwords are passed in urls) or optimal to use. However, it works. And it is nice.

## Idea:
* One server instance can hold data of multiple users, in case of crash or
 logout all data is stored on the disk
* Supports hash object which is a hash of hashes from string to string and Pairs object which is just a string hash
* Server communicates with client over http
* GET requests represent data fetch operations
* POST requests represent data store operations. Data is send as json in body of a request 
* DELETE requests represent data deletion operations

## Api:
* register by issuing get request at `/reg/{username}:{password}`
* quit by issuing get at `/quit/{username}:{password}`
* perform operation on hash issuing requests to `{username}:{password}/hash`:
    * `...hash/key` - access value at given key of your hash
    * `...hash/key1/key2` - access value at key1 of a hash stored at key2
    * `...hash` - issue GET or DELETE to fetch all data or delete all data
* perform operation on pairs issuing requests to `{username}:{password}/pairs`:
    * `...pairs/key` - access value at given key of your pairs object
    * `...pairs` - issue GET or DELETE to fetch all data or delete all data

## Web client for this application is in development...