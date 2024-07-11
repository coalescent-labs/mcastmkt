# mcastmkt
Simple tool to test multicast traffic flows and to get a better understanding of the market protocols.

## Description

This is an example project on how to build a command line tool with Go using the Viper and Cobra libraries.

It is a simple tool to test multicast traffic flows and to get a better understanding 
of the market protocols and that can be easily extended to support other multicast flows.

It also can be used to test the network configuration and the multicast traffic flow in a network
by sending and receiving multicast packets with the `any/send` and `any/listen` commands.

## Build

You can simply build the mcastmkt project with the provided Makefile.

```
# build the binary
make build

# build the multi arch binaries (default linux and windows)
make clean build_all

# build the multi arch binaries with specific version
VERSION=0.0.1 make clean build_all
```

See the makefile for details.

## Usage

The following manual is generated by cobra help output:

```
mcastmkt –  command-line tool to test markets multicast traffic flows

Usage:
  mcastmkt [command]

Available Commands:
  any         Generic multicast commands without enter in specific market protocol and conversion
  completion  Generate the autocompletion script for the specified shell
  eurex       Eurex multicast commands
  euronext    Euronext optiq multicast commands
  help        Help about any command

Flags:
  -c, --config string   config file (default is $HOME/.mcastmkt.yaml)
  -h, --help            help for mcastmkt
  -v, --version         version for mcastmkt

Use "mcastmkt [command] --help" for more information about a command.
```

Examples:
```
# Listen to multicast traffinc and dump any received packed to stdout
mcastmkt any listen -a 224.0.212.78:40078 -i eno1 -d

# Send muldicast test packets...
mcastmkt any send -a 224.50.50.59:59001 -d -i 192.168.178.128 -t 1 -n 5000
# ... and listen to them
mcastmkt any listen -a 224.50.50.59:59001 -d -i 192.168.178.128

# Listen to Eurex EMDI multicast traffic and dump out of sequence or duplicates messages
mcastmkt eurex listen emdi -a 224.0.50.59:59001 -i eno1
# As previous but it also dumps all received packets to stdout
mcastmkt eurex listen emdi -a 224.0.50.59:59001 -i eno1 -d

# Listen to Euronext Optiq MDG multicast stream and dump out of sequence or duplicates messages
mcastmkt euronext listen mdg -a 224.0.212.78:40078 -i eno1
```