# snap collector plugin - Cassandra

This plugin collects Cassandra cluster statistics with the [Snap Framework] (http://github.com/intelsdi-x/snap).


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

* [Snap](http://github.com/intelsdi-x/snap)
* Cassandra node/cluster
* [golang 1.5+](https://golang.org/dl/)
* [snap-plugin-utilities](http://github.com/intelsdi-x/snap-plugin-utilities)

Note that Go and plugin utilities are needed only if building the plugin from source.

### Operating systems
All OSs currently supported by Snap:
* Linux/amd64
* Darwin/amd64

### Installation

#### Download plugin binary:

You can also download prebuilt binaries for OS X and Linux (64-bit) at the [releases](https://github.com/intelsdi-x/snap-plugin-collector-cassandra/releases) page

#### To build the plugin binary:
Get the source by running a `go get` to fetch the code:
```
$ go get -d github.com/intelsdi-x/snap-plugin-collector-cassandra
```

Build the plugin by running make within the cloned repo:
```
$ cd $GOPATH/src/github.com/intelsdi-x/snap-plugin-collector-cassandra && make
```
This builds the plugin in `/build/rootfs/`


### Configuration and Usage

* Set up the [Snap framework](https://github.com/intelsdi-x/snap/blob/master/README.md#getting-started).

* Add Cassandra configuration information to the Task Manifest (see [examples](#examples))

* Load the plugin and create the task

## Documentation 

### Collected Metrics
This plugin has the ability to gather all metrics within the Cassandra org.apache.cassandra.metrics package. View [Metric Types](./METRICS.md) for the full list. They are in the following catalog:

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

The dynamic metric queries are supported. You may view the [sample dynamic metrics](./DYNAMIC_METRICS.md).

### Examples
Example running snap-plugin-collector-collector, passthru processor, and writing data to a file. User need to provide following parameters in the global configuration of the collector.

* `"url"` - The domain URL of cassandra server (ex. `"192.168.99.100"`)
* `"port"` - The port number of Cassandra MX4J (ex. `"8082"`)

Refer to [Sample Gloabal Configuration](./examples/cfg/cfg.json).

*Optional:* Run Cassandra from docker image:
```
docker run --detach --name snap-cassandra -p 9042:9042 -p 7199:7199 -p 8082:8082 -p 9160:9160 -d candysmurfhub/cassandra
```
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
```
/intel/cassandra/node/*/org_apache_cassandra_metrics/type/*/scope/*/name/*/999thPercentile 		 3
/intel/cassandra/node/*/org_apache_cassandra_metrics/type/*/scope/*/name/*/99thPercentile 		 3
/intel/cassandra/node/*/org_apache_cassandra_metrics/type/*/scope/*/name/*/Count 			 3
/intel/cassandra/node/*/org_apache_cassandra_metrics/type/*/scope/*/name/*/FifteenMinuteRate 		 3
/intel/cassandra/node/*/org_apache_cassandra_metrics/type/*/scope/*/name/*/FiveMinuteRate 		 3
/intel/cassandra/node/*/org_apache_cassandra_metrics/type/*/scope/*/name/*/Max
/intel/cassandra/node/*/org_apache_cassandra_metrics/type/*/keyspace/*/scope/*/name/*/50thPercentile 		 3
/intel/cassandra/node/*/org_apache_cassandra_metrics/type/*/keyspace/*/scope/*/name/*/75thPercentile 		 3
```

$SNAP_PATH/bin/snapctl metric list --verbose
```
/intel/cassandra/node/[Node Name]/org_apache_cassandra_metrics/type/[typeValue]/scope/[scopeValue]/name/[nameValue]/FiveMinuteRate 		 		 3 		 double
/intel/cassandra/node/[Node Name]/org_apache_cassandra_metrics/type/[typeValue]/scope/[scopeValue]/name/[nameValue]/Max 			 		 3 		 double
/intel/cassandra/node/[Node Name]/org_apache_cassandra_metrics/type/[typeValue]/scope/[scopeValue]/name/[nameValue]/Mean 			 		 3 		 double
/intel/cassandra/node/[Node Name]/org_apache_cassandra_metrics/type/[typeValue]/scope/[scopeValue]/name/[nameValue]/MeanRate 		 			 3 		 double
/intel/cassandra/node/[Node Name]/org_apache_cassandra_metrics/type/[typeValue]/scope/[scopeValue]/name/[nameValue]/Min 			 		 3 		 double
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
                "/intel/cassandra/node/*/org_apache_cassandra_metrics/type/*/keyspace/*/name/*/Value":{},
                "/intel/cassandra/node/*/org_apache_cassandra_metrics/type/*/keyspace/*/scope/*/name/*/50thPercentile": {},
                "/intel/cassandra/node/*/org_apache_cassandra_metrics/type/*/keyspace/*/scope/*/name/*/Max":{}
            },
            "config": {
                "/intel/cassandra": {
                    "url": "192.168.99.100",
                    "port": 8082
                }
            },
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
                    ]
                }
            ]
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
```
2016-06-04 10:34:38.477422325 -0700 PDT|/intel/cassandra/node/egu-mac01.lan/org.apache.cassandra.metrics/type/Table/keyspace/system_traces/scope/sessions/name/RangeLatency/Max|0
2016-06-04 10:34:38.477423472 -0700 PDT|/intel/cassandra/node/egu-mac01.lan/org.apache.cassandra.metrics/type/Table/keyspace/system_traces/scope/sessions/name/CasProposeLatency/Max|0
2016-06-04 10:34:38.47742462 -0700 PDT|/intel/cassandra/node/egu-mac01.lan/org.apache.cassandra.metrics/type/Table/keyspace/system_traces/scope/sessions/name/TombstoneScannedHistogram/Max|0
2016-06-04 10:34:38.477425708 -0700 PDT|/intel/cassandra/node/egu-mac01.lan/org.apache.cassandra.metrics/type/Table/keyspace/system_traces/scope/sessions/name/ReadLatency/Max|0
2016-06-04 10:34:38.477426822 -0700 PDT|/intel/cassandra/node/egu-mac01.lan/org.apache.cassandra.metrics/type/Table/keyspace/system_traces/scope/sessions/name/ViewReadTime/Max|0
2016-06-04 10:34:38.477427958 -0700 PDT|/intel/cassandra/node/egu-mac01.lan/org.apache.cassandra.metrics/type/Table/keyspace/system_traces/scope/sessions/name/CoordinatorReadLatency/Max|0
2016-06-04 10:34:38.477439374 -0700 PDT|/intel/cassandra/node/egu-mac01.lan/org.apache.cassandra.metrics/type/Table/keyspace/system_traces/sc
```

### Roadmap
This plugin is still in active development. As we launch this plugin, we have a few items in mind for the next few releases:
- [ ] Additional error handling
- [ ] Cluster load and scalability testing

If you have a feature request, please add it as an [issue](https://github.com/intelsdi-x/snap-plugin-collector-cassandra/issues) 
and feel free to submit a [pull request](https://github.com/intelsdi-x/snap-plugin-collector-cassandra/pulls) that resolves it.

## Community Support
This repository is one of **many** plugins in **Snap**, the open telemetry framework. See the full project at http://github.com/intelsdi-x/snap. To reach out to other users, head to the [main framework](https://github.com/intelsdi-x/snap#community-support).

## Contributing
We love contributions!

There's more than one way to give back, from examples to blogs to code updates. See our recommended process in [CONTRIBUTING.md](CONTRIBUTING.md).

And **thank you!** Your contribution, through code and participation, is incredibly important to us.

## License
[Snap](http://github.com:intelsdi-x/snap), along with this plugin, is an Open Source software released under the Apache 2.0 [License](LICENSE).


## Acknowledgements

* Author: [@candysmurf](https://github.com/candysmurf/)

