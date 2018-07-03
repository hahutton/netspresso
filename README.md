Netspresso
==========

Personal load testing and latency measurement for CosmosDB

Not a good example ;)

This creates synthetic load on CosmosDB. It PUTs synthetic objects to
a single write regions, stores the newly generated key in redis for
subsequent use in querying from instances across the globe. 

The results of all requests are measured for latency. They are then
packaged into JSON event blobs and pushed into and EventHub. 

The events are best consumed by Time Series Insights to visualize
latencies and throughputs of CosmosDB from across the world.



