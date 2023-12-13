# Archway blockchain data publisher

[![Latest release](https://img.shields.io/github/v/release/SyntropyNet/archway-publisher)](https://github.com/SyntropyNet/archway-publisher/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Getting Started

### Prerequisites

Before you start using (modifying) Archway blockchain data publisher locally, there are some prerequisites you need to fulfill:

* Access to Syntropy Data Layer and its streams - [Developer Portal](https://developer-portal.syntropynet.com/)
  * A Registered Publisher Profile
  * Generated User JWT token and NKey Seed
* Go (Golang 1.20) - [Install Go](https://go.dev/doc/install)
* Access to Archway full node

NOTE: It is possible to run Archway publisher without a connection to broker. This way a Stub will be used that simply logs messages that would be sent to the DL Broker otherwise.

### Building

In order to build the binary, you can simply run:

```bash
make build
```

The executable binary will be stored inside `./dist` directory. You can explore more Makefile targets by running this command:

```bash
make help
```

In order to build the Docker container, you can run this command:

```bash
make docker-build
```

### Quick start

The publisher requires these parameters to start:

* Data Layer Broker address
* Broker NKey Seed (NATS User NKey)
* Broker User JWT token
* Archway full node Tendermint endpoint address (by default 26657 port)
* Archway full node RPC address (by default 1317 port)
* Archway full node gRPC address (by default 9090 port)
* Optional Archway publisher subject prefix name

The most convenient way to setup these parameters is by creating an `.env` file in the working directory of the binary, e.g.:

```bash
NATS_URL=<Broker Address>
NATS_NKEY=SUAMBOK<...>
NATS_JWT=eyJ0eX<...>
ARCHWAY_TENDERMINT=tcp://127.0.0.1:26657
ARCHWAY_RPC=http://127.0.0.1:1317
ARCHWAY_GRPC=127.0.0.1:9090
ARCHWAY_PREFIX=archway-sandbox
```

Then, running the `./build/archway-publisher` binary would start the publisher and begin publishing.

### Configuring Archway Publisher as systemd service

Assuming these prerequisites are met:

* Archway full node is running locally using cosmovisor
* Archway publisher will be running as `archway` user with `/home/archway` as home directory
* `.env` file is in `archway` home directory

First, one has to create the service definition file in `/etc/systemd/system/archway-publisher.service`:

```bash
[Unit]
Description=Archway publisher service
After=cosmovisor.service

[Service]
User=archway
Environment="LD_LIBRARY_PATH=/usr/local/lib"
WorkingDirectory=/home/archway
ExecStart=/home/archway/archway-publisher start
Restart=always
RestartSec=3
LimitNOFILE=infinity
LimitNPROC=infinity

[Install]
WantedBy=multi-user.target
```

Enable and start the service:

```bash
sudo systemctl enable archway-publisher
sudo systemctl start archway-publisher
```

You can monitor Archway publisher logs using this command:

```bash
sudo journalctl -f -u archway-publisher
```

### Resolving libwasmvm dependency

Archway leverages WASM heavily for its features. Therefore it depends on WASM library.
In case the binary was built on one machine while the binary is meant to run on another machine, you must
copy `libwasmvm.x86_64.so`(if the target architecture is x86_64) to `/usr/local/lib` on the target machine.

You can find the path to this library by using `ldd` tool like this:

```bash
ldd dist/archway-publisher 
```

Then the output will contain something like this (look for a line with `libwasmvm` - the exact file will depend on your host architecture):

```bash
        linux-vdso.so.1 (0x00007ffe129c0000)
        libresolv.so.2 => /lib/x86_64-linux-gnu/libresolv.so.2 (0x00007f218ac6a000)
        libpthread.so.0 => /lib/x86_64-linux-gnu/libpthread.so.0 (0x00007f218ac47000)
        libwasmvm.x86_64.so => <path to user home directory>/go/pkg/mod/github.com/!cosm!wasm/wasmvm@v1.4.1/internal/api/libwasmvm.x86_64.so (0x00007f218a52b000)
        libc.so.6 => /lib/x86_64-linux-gnu/libc.so.6 (0x00007f218a339000)
        /lib64/ld-linux-x86-64.so.2 (0x00007f218ac94000)
        libgcc_s.so.1 => /lib/x86_64-linux-gnu/libgcc_s.so.1 (0x00007f218a31e000)
        librt.so.1 => /lib/x86_64-linux-gnu/librt.so.1 (0x00007f218a312000)
        libm.so.6 => /lib/x86_64-linux-gnu/libm.so.6 (0x00007f218a1c3000)
        libdl.so.2 => /lib/x86_64-linux-gnu/libdl.so.2 (0x00007f218a1bd000)
```

The provided `Makefile` will automatically find this library and copy to `dist` folder.

### Archway full node configuration

Default Archway RPC body size limit prevents some of the messages to be delivered to the publisher. Therefore
Archway full node `config.toml` should be modified like so:

```config
<...>
#######################################################
###       RPC Server Configuration Options          ###
#######################################################
[rpc]
<...>
experimental_subscription_buffer_size = 80000
<...>
experimental_websocket_write_buffer_size = 80000
<...>
experimental_close_on_slow_client = true
<...>
max_body_bytes = 10000000

```

## Future development

As this is a barebones Archway blockchain data publisher, there is a lot of room to improve.

## Contributing

We welcome contributions from the community. Whether it's a bug report, a new feature, or a code fix, your input is valued and appreciated.

## Syntropy

If you have any questions, ideas, or simply want to connect with us, we encourage you to reach out through any of the following channels:

* **Discord**: Join our vibrant community on Discord at [https://discord.com/invite/jqZur5S3KZ](https://discord.com/invite/jqZur5S3KZ). Engage in discussions, seek assistance, and collaborate with like-minded individuals.
* **Telegram**: Connect with us on Telegram at [https://t.me/SyntropyNet](https://t.me/SyntropyNet). Stay updated with the latest news, announcements, and interact with our team members and community.
* **Email**: If you prefer email communication, feel free to reach out to us at <devrel@syntropynet.com>. We're here to address your inquiries, provide support, and explore collaboration opportunities.

## License

This project is licensed under the terms of the MIT license.
