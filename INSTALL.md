# JENS-CLI
This package contains commands for playing a data rate pattern on a network interface which uses a l4s capable queue.
Link capacity is changed over time in uplink direction in a fix time period (frequency up to 100 Hz), 
measures of the state of the l4s queue are sampled (10ms) and can be stored (csv or psql). 
The data rate pattern is defined in a csv file, an example is provided at /etc/jens-cli/drp_3valleys.csv.
The command drshow visualizes measures or data rate patterns on a terminal ui.

**The Package provides tree executables (`drplay` & `drshow` & `drbenchmark`)**

# Installation

## Prerequisites
* Deactivate `secure boot` option in the BIOS
* Running Debian system (version: `bullseye`)
* Installation of docker (if grafana/postgres container should run)

## Installation on a Debian-System:
**jens-cli runs on debian linux, it requires an update of kernel packages. This means it cannot be used from within a docker-container.**

**But it can be installed in a VM. following the same steps**

`Root` privileges are required

* Make the JENS package repository visible by executing the script `scripts/setupApt.sh`
* Run `apt install jens-cli`
* The executables drplay drbenchmark and drshow are now ready to be used
  `drplay -dev eno1 -loop | drshow`
* Optionally start the Grafana Dashboard by 
  `docker run -d --net=host --restart always --name drdashboard edgeaml/drdash`
* Configure your networking to route traffic through your device, maybe configure the wifi network adapter as an access point. 
  Alternatively you can start your L4S sender on the device.

--------

# Usage

## drplay
The parameter -h gives information about execution parameters. Also the manpages contain information about this tool.

### start the data rate player
To launch the data rate player (drplay) you can invoke it using `drplay <arguments>`.

Autocomplete (pressing tab) will help giving all needed parameters.

_drplay has to be run as root (`sudo -i`) or user with privileges for tc._
* e.g. play data rate pattern on nic lo with frequency of 10 Hz in loop mode, l4s queue measures are written to stdout
  * `drplay -dev lo -pattern /etc/jens-cli/drp_3valleys.csv -freq 10 -loop`

Output e.g.:

```
timestampMs sojournTimeMs loadKbits capacityKbits ecnCePercent dropped prio netflow
1678808847209 0 100 13360 0 0 0 2 192.168.178.119:22-192.168.178.75:57442
1678808848179 0 400 13360 0 0 0 2 192.168.178.119:40184-192.168.178.75:5201
1678808848179 0 200 13650 0 0 0 2 192.168.178.119:22-192.168.178.75:57442
1678808848179 0 0 13940 0 0 0 2 192.168.178.119:60970-192.168.178.75:5201
1678808848229 0 13100 13940 0 0 0 2 192.168.178.119:60970-192.168.178.75:5201
1678808848229 0 0 13940 0 0 0 2 192.168.178.119:40184-192.168.178.75:5201
1678808848240 1 13100 13951 0 0 0 2 192.168.178.119:60970-192.168.178.75:5201
...
```

### output measures
- `timestampMs` : Epoch timestamp in ms of the sample
- `sojournTimeMs` : Average time an ip packet stayed in the queue in ms for sample
- `loadKbits` : Load in Kbits for sample
- `capacityKbits`: Capacity set by the data rate pattern. (minimum is set in config)
- `ecnCePercent` : Percentage of ip packets with ECN=CE in sample
- `dropped`: Number of packets dropped in sample
- `prio`: internal priority of the queue, in which the packet is sent: 1=high, 2=medium, 3=low
- `srcIp`: source ip of ip packet
- `dstIp`: destination ip of ip packet
- `netflow`: srcIp:srcPort-dstIp:dstPort

### Configuration
The Config file can be used to adjust certain parameters, that are not configurable through the cmdl arguments. Such as the static addon-latency 


## drshow
Drshow is a visualization utility.

For help regarding this command see `man drshow` and `drshow --help`.

Drshow (`drshow`) can be used in two modes:

### Mode 1: PipeMode
Additional Help: `drshow -h pipe`

In PipeMode it is used in conjunction with `drplay`. The stdout Output is piped into `drshow` to visualize it.
Inside of the program the Arrowkeys (Up and Down) can be used to navigate the flows.
Various Stats are displayed in a human readable UI.
```sh
drplay [opts] | drshow
```
#### Note
`drshow` might not react to Ctrl-C in Pipemode if there is no traffic on the interface used together with `drplay`. In this case, for now, close and reopen the terminal window. 

### Mode 2: StaticMode
Additional Help: `drshow -h static`

When `drshow` is used without a piped in input, it needs a directory or single file to display.
When a directory is choosen, the user can navigate the .csv files and inspect them.
```sh
drshow patterns/mydrp.csv
```
## drbenchmark
`drshow` is tool to launch drplay multiple times with set configurations 

For help regarding this command see `man drbenchmark` or `drbenchmark --help`.

`drbenchmark` uses JSON configurations to define a Benchmark.

```json
{
  "Hash":"",
  "Inner":{
    "Name":"Quick Example",
    "MaxBitrateEstimationTimeS":8,
    "Patterns":[
      {
        "Path": "/etc/jens-cli/drp_3valleys.csv",
        "Hash": ""
      },{
        "Path": "/etc/jens-cli/drp_munich_village.csv",
        "Hash": "",
        "Setting": {
          "DRP":{
            "Scale": 0.8
          }
        }
      }
    ],
    "DrplaySetting":{
      "DRP":{
        "Frequency": 10
      },
      "TC": {
        "Markfree": 2,
        "Markfull": 4,
        "Queuesizepackets": 10000,
        "Extralatency": 10
      }
    }
  }
}
```

### Json Explanation:

- `Hash`, hash of `Inner`
- `Inner`:
    - `Name`, benchmark name. Used for human identification
    - `Max_application_bitrate`, is not currently in use. Might get removed
    - `Patterns`:
        - `Path`, filepath
        - `Hash`, hash of the pattern
        - `Setting`: 
            - `DRP`, for drplay
            - `TC`, for drplay
    - `DrplaySetting`: 
      - `DRP`, for drplay
      - `TC`, for drplay

All Hash Values can be set to `""`. Only if they are set a comparison is made.

The file and consequently the benchmark is read and for not defined settings, fallbacks are used. This results in the following order of config-values: `1. json.patterns.Setting > 2. json.DrPlaySetting > 3. config.toml`


# ConfigFile
The config file contains some settings for tc commands, `drplay`, `drshow`, `drbenchmark` and the connection to the PorstgeSQL server.
The config file is located in `/etc/jens-cli/config.toml`.

```toml
[tccommands]
  markfree=4 #in Millisecond
  markfull=14 #in Millisecond
  # Size of custom Queue in packets
  queuesizepackets=10000
  # AddonLatency in MS to add to all packets
  extralatency=20
  # Mark non-ect(1) Traffic as enabled
  l4sEnabledPreMarking=false
  # set queue priority handling of packets: low, medium or high
  # qosmode=0: IPTOS_LOWDELAY increase packet priority in queue, IPTOS_THROUGHPUT decrease
  # qosmode=1: Any IPv6 and IPv4 traffic is sorted into the normal
  # qosmode=2: L4S mode, only ECT(1) ip packets get into high priority
  qosmode=2
  # Mark the first packets with special ect
  signalDrpStart=false

[postgres]
  dbname = "l4s_measure"
  host = "localhost"
  password = "changeDefaultPassword"
  port = 5432
  user = "edge"

[drp]
  # MinimumDataRate for patterns
  minRateKbits = 500
  # Phase in MS before Networkshaping /DRP takes effect
  WarmupBeforeDrpMs = 2000

[drshow]
  scalePlots=true #instead of scrolling
  exportPath="/etc/jens-cli"
  #Used for filtering of flows in pipemode
  FilterLevel0Until=7       
  #Adds to Level0Until: Min count of flows for filter to kick in action
  FilterLevel1AddUntil=6    
  #Filter minium samples to display flow
  FilterLevel1MinSamples=15 
  #Or maximum time passed since last sample was received
  FilterLevel1SecsPassed=30 
  (...)


```
An example data-rate-pattern can also be found in `/etc/jens-cli`.

# Dashboard
An exemplary Grafana-Instance together with a PostrgresSQL instance configured to use with the jens-cli can be found on dockerhub.

## Login

Username: `admin`

Password: `changeDefaultPassword`

## Running

`docker run -d --net=host --name drdash --restart unless-stopped edgeaml/drdash:latest`

If docker was installed without root-privileges, you need to expose the ports of both grafana (`3000`) and psql (`5432`) using the `-p` parameter in the docker run command instead of `--net=host`.

# Annex

## Sample Test-Setup

### Requirements:
1. SSH connection to JENS (x2)
  1. Iperf3 installed
2. Other machine connected to JENS or in the same network 
  1. Iperf3 installed

### Setup:

#### On the other machine
1. Launch iperf3 as server (`iperf3 -s`)

#### On JENS / the ssh connection to JENS
* Launch `drplay -dev <dev> -psql -tag iperf3-setup-test-ecn1`
  * With `<dev>` being a NIC on JENS (e.g. `wlp0s20f3` or `eno1`or ... )
* Launch iperf3 as a Client `iperf3 -c <other_machine_ip> --udp -t 100 -i 0.5 -b 18M -S 0x1`
  * `-S 0x1` Mark the packets as ecn-enabled
  * `-t 100` Send data for 100 seconds
