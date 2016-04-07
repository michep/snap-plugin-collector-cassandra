# snap collector plugin - Cassandra

This plugin collects Cassandra cluster statistics by using snap telemetry engine.

The intention for this plugin is to collect metrics for Cassandra nodes and cluster health.

This plugin is used in the [snap framework] (http://github.com/intelsdi-x/snap).


1. [Getting Started](#getting-started)
  * [System Requirements](#system-requirements)
  * [Operating systems](#openrating-systems)
  * [Installation](#installation)
  * [Configuration and Usage](#configuration-and-usage)
2. [Documentation](#documentation)
  * [Collected Metrics](#collected-metrics)
  * [Examples](#examples)
  * [Roadmap](#roadmap)
3. [Community Support](#community-support)
4. [Contributing](#contributing)
5. [License](#license)
6. [Acknowledgements](#acknowledgements)

## Getting Started

In order to use this plugin you need the Cassandra node or cluster that you can collect metrics from.

### System Requirements

* [snap](http://github.com/intelsdi-x/snap)
* [snap-plugin-utilities](http://github.com/intelsdi-x/snap-plugin-utilities)
* Cassandra node/cluster
* [golang 1.5+](https://golang.org/dl/)

Note that Golang is needed only if building the plugin from the source.

### Operating systems
All OSs currently supported by snap:
* Linux/amd64
* Darwin/amd64

### Installation

### Install Cassandra from docker image
```
docker run --detach --name snap-cassandra -p 9042:9042 -p 7199:7199 -p 8082:8082 -p 9160:9160 -d candysmurfhub/cassandra
```

#### To build the plugin binary:
Get the source by running a `go get` to fetch the code:
```
$ go get github.com/intelsdi-x/snap-plugin-collector-cassandra
```

Build the plugin by running make within the cloned repo:
```
$ cd $GOPATH/src/github.com/intelsdi-x/snap-plugin-collector-cassandra && make
```
This builds the plugin in `/build/rootfs/`

#### Builds
You can also download prebuilt binaries for OS X and Linux (64-bit) at the [releases](https://github.com/intelsdi-x/snap-plugin-collector-cassandra/releases) page

### Configuration and Usage
* Set up the [snap framework](https://github.com/intelsdi-x/snap/blob/master/README.md#getting-started)
* Ensure `$SNAP_PATH` is exported  
`export SNAP_PATH=$GOPATH/src/github.com/intelsdi-x/snap/build`
* Ensure the global configuration is provided. See [example](#examples) for details.

## Documentation

To learn more about this plugin:

* [snap cassandra examples](#examples)

### Collected Metrics
This plugin has the ability to gather all metrics within cassandra org.apache.cassandra.metrics packge.
View [Metric Types](./METRICS.md). They are in the following catalog:

**Cassandra Metric Catalog**
* Cache (Counter, Row, Key)
* Client Request (CASRead, CASWrite, Read, Write)
* ColumnFamily (system_auth, system, ...)
* CommitLog
* Compaction 
* CQL
* Dropped Message
* File Cache
* Keyspace
* Storage
* Thread Pool

### Examples
Example running snap-plugin-collector-collector, passthru processor, and writing data to a file.
User need to provide following parameters in the global configuration of the collector.

* `"url"` - The domain URL of cassandra server (ex. `"192.168.99.100"`)
* `"port"` - The port number of Cassandra MX4J (ex. `"8082"`)
* `"hostname"` - The host name of Cassandra

Refer to [Sample Gloabal Configuration](./examples/cfg/cfg.json)

![Docker example](https://media.giphy.com/media/3osxY4vXZSyxYZG8eY/giphy.gif)

In one terminal window, open the snap daemon (in this case with logging set to 1 and trust disabled):
```
$ $SNAP_PATH/bin/snapd -l 1 -t 0 --config <path to global cfg.json>
```
In another terminal window:
Load snap-plugin-collector-cassandra
```
$ $SNAP_PATH/bin/snapctl plugin load <path to snap-plugin-collector-cassandra>
Plugin loaded
Name: cassandra
Version: 1
Type: collector
Signed: false
Loaded Time: Thu, 10 Mar 2016 22:31:34 PST
```
See available metrics for your system (this is just part of the list)
```
$SNAP_PATH/bin/snapctl metric list                                
NAMESPACE 												 VERSIONS
/intel/Cassandra/Node/egu-mac01.lan/org.apache.cassandra.metrics/type/ThreadPools/path/internal/scope/InternalResponseStage/name/MaxPoolSize/Value 					 1
/intel/Cassandra/Node/egu-mac01.lan/org.apache.cassandra.metrics/type/ThreadPools/path/internal/scope/InternalResponseStage/name/PendingTasks/Value 					 1
/intel/Cassandra/Node/egu-mac01.lan/org.apache.cassandra.metrics/type/ThreadPools/path/internal/scope/InternalResponseStage/name/TotalBlockedTasks/Count 				 1
/intel/Cassandra/Node/egu-mac01.lan/org.apache.cassandra.metrics/type/ThreadPools/path/internal/scope/MemtableFlushWriter/name/ActiveTasks/Value 					 1
/intel/Cassandra/Node/egu-mac01.lan/org.apache.cassandra.metrics/type/ThreadPools/path/internal/scope/MemtableFlushWriter/name/CompletedTasks/Value 					 1
/Intel/Cassandra/Node/egu-mac01.lan/org.apache.cassandra.metrics/type/ThreadPools/path/internal/scope/MemtableFlushWriter/name/CurrentlyBlockedTasks/Count 				 1
/Intel/Cassandra/Node/egu-mac01.lan/org.apache.cassandra.metrics/type/ThreadPools/path/internal/scope/MemtableFlushWriter/name/MaxPoolSize/Value 					 1
/Intel/Cassandra/Node/egu-mac01.lan/org.apache.cassandra.metrics/type/ThreadPools/path/internal/scope/MemtableFlushWriter/name/PendingTasks/Value 					 1
/Intel/Cassandra/Node/egu-mac01.lan/org.apache.cassandra.metrics/type/ThreadPools/path/internal/scope/MemtableFlushWriter/name/TotalBlockedTasks/Count 					 1
```

Load passthru plugin for processing:
```
$SNAP_PATH/bin/snapctl plugin load $SNAP_PATH/plugin/snap-processor-passthru
Plugin loaded
Name: passthru
Version: 1
Type: processor
Signed: false
Loaded Time: Thu, 10 Mar 2016 22:33:45 PST
```

Load file plugin for publishing:
```
$SNAP_PATH/bin/snapctl plugin load $SNAP_PATH/plugin/snap-publisher-file  
Plugin loaded
Name: file
Version: 3
Type: publisher
Signed: false
Loaded Time: Thu, 10 Mar 2016 22:34:28 PST
```

Create a task manifest file (e.g. `cassandra-collector-task.json`. replace node id):    
```json
{
    "version": 1,
    "schedule": {
        "type": "simple",
        "interval": "1s"
    },
    "workflow": {
        "collect": {
            "metrics": {
                "/intel/cassandra/node/egu-mac01.lan/org.apache.cassandra.metrics/type/ThreadPools/path/internal/scope/ValidationExecutor/name/MaxPoolSize/*": {},
                "/intel/cassandra/node/egu-mac01.lan/org.apache.cassandra.metrics/type/Keyspace/keyspace/system/name/TombstoneScannedHistogram/*": {},
                "/intel/cassandra/node/egu-mac01.lan/org.apache.cassandra.metrics/type/Keyspace/keyspace/system/name/BloomFilterOffHeapMemoryUsed/*": {},
                "/intel/cassandra/node/egu-mac01.lan/org.apache.cassandra.metrics/type/Keyspace/keyspace/system_auth/name/ReadLatency/*": {},
                "/intel/cassandra/node/egu-mac01.lan/org.apache.cassandra.metrics/type/DroppedMessage/scope/READ/name/Dropped/*": {}
            },
            "config": {},
            "process": [
                {
                    "plugin_name": "passthru",
                    "process": null,
                    "publish": [
                        {
                            "plugin_name": "file",
                            "config": {
                                "file": "/tmp/collected_cassandra"
                            }
                        }
                    ],
                    "config": null
                }
            ],
            "publish": null
        }
    }
}
```

Create task:
```
$SNAP_PATH/bin/snapctl task create -t ~/task/cassandra-collector-task.json
Using task manifest to create task
Task created
ID: 713bb201-eb69-407a-b11e-0d6606a444ca
Name: Task-713bb201-eb69-407a-b11e-0d6606a444ca
State: Running
```

See file output (this is just part of the file):
```
$ tail -f collected_cassandra
2016-03-10 22:38:14.690001495 -0800 PST|[org.apache.cassandra.metrics type DroppedMessage scope READ name Dropped Count]|0|egu-mac01.lan
2016-03-10 22:38:14.690018347 -0800 PST|[org.apache.cassandra.metrics type DroppedMessage scope READ name Dropped FiveMinuteRate]|0|egu-mac01.lan
2016-03-10 22:38:14.690030047 -0800 PST|[org.apache.cassandra.metrics type DroppedMessage scope READ name Dropped OneMinuteRate]|0|egu-mac01.lan
2016-03-10 22:38:14.690118077 -0800 PST|[org.apache.cassandra.metrics type DroppedMessage scope READ name Dropped FifteenMinuteRate]|0|egu-mac01.lan
2016-03-10 22:38:14.690130891 -0800 PST|[org.apache.cassandra.metrics type DroppedMessage scope READ name Dropped MeanRate]|0|egu-mac01.lan
2016-03-10 22:38:14.695199595 -0800 PST|[org.apache.cassandra.metrics type Keyspace keyspace system name BloomFilterOffHeapMemoryUsed Value]|352|egu-mac01.lan
2016-03-10 22:38:14.700835298 -0800 PST|[org.apache.cassandra.metrics type Keyspace keyspace system name TombstoneScannedHistogram 50thPercentile]|1|egu-mac01.lan
2016-03-10 22:38:14.700882825 -0800 PST|[org.apache.cassandra.metrics type Keyspace keyspace system name TombstoneScannedHistogram 95thPercentile]|1|egu-mac01.lan
2016-03-10 22:38:14.700962157 -0800 PST|[org.apache.cassandra.metrics type Keyspace keyspace system name TombstoneScannedHistogram 999thPercentile]|1|egu-mac01.lan
2016-03-10 22:38:14.700990817 -0800 PST|[org.apache.cassandra.metrics type Keyspace keyspace system name TombstoneScannedHistogram Count]|36|egu-mac01.lan
2016-03-10 22:38:14.701016028 -0800 PST|[org.apache.cassandra.metrics type Keyspace keyspace system name TombstoneScannedHistogram Mean]|1|egu-mac01.lan
```

### Roadmap
This plugin is still in active development. As we launch this plugin, we have a few items in mind for the next few releases:
- [ ] Additional error handling
- [ ] Cluter load and scalability testing

If you have a feature request, please add it as an [issue](https://github.com/intelsdi-x/snap-plugin-collector-cassandra/issues) 
and/or submit a [pull request](https://github.com/intelsdi-x/snap-plugin-collector-cassandra/pulls).

## Community Support
This repository is one of **many** plugins in the **snap**, a powerful telemetry agent framework. See the full project at 
http://github.com/intelsdi-x/snap. To reach out to other users, head to the [main framework](https://github.com/intelsdi-x/snap#community-support).


## Contributing
We love contributions!

There is more than one way to give back, from examples to blogs to code updates.

## License

[snap](http://github.com/intelsdi-x/snap), along with this plugin, is an Open Source software released under the Apache 2.0 [License](LICENSE).


## Acknowledgements

* Author: [@candysmurf](https://github.com/candysmurf/)

