# tendermint-testing
Testing tendermint with netrix

Should be run with forked tendermint which is integrated to run with netrix. Can be found [here](https://github.com/zeu5/tendermint/tree/pct-instrumentation)

## Getting started
### Requirements
- Golang 1.18
- Docker
- [docker-compose](https://pypi.org/project/docker-compose/). Not the one built-in to recent versions of docker. I found pip to be the easiest way to install it: `pip3 install --user docker-compose`.


### Instrumented tendermint nodes
The tests require a modified version of tendermint found [here](https://github.com/zeu5/tendermint/tree/pct-instrumentation). 
Make sure to download the `pct-instrumentation` branch.

```shell
# Clone somewhere outside of this repository
cd ../
git clone git@github.com:zeu5/tendermint.git tendermint-pct-instrumentation
git checkout pct-instrumentation
```

The nodes will run in a docker-compose setup as documented [here](https://github.com/tendermint/tendermint/blob/master/docs/tools/docker-compose.md), and will communicate with your testing server running on the host.

To this end, docker-compose will create a bridge network for you when you first start the nodes. 
On linux, the default IP assigned to the host will be `192.167.0.1`. It may be different on your machine.
Make sure this IP is set in the nodes config file, or they will be unable to reach the testing server:

```toml
# ../tendermint-pct-instrumentation/networks/local/localnode/config-template.toml

controller-master-addr = "192.167.0.1:7074"
```

Now under `tendermint-pct-instrumentation`:

```shell
# Build the linux binary in ./build
make build-linux

# Build tendermint/localnode image
make build-docker-localnode

# Create the bridge network
docker-compose up --no-start
```

Check your host IP on the bridge network:

```shell
$ ip addr
...
4: br-48dbf8e89480: <NO-CARRIER,BROADCAST,MULTICAST,UP> mtu 1500 qdisc noqueue state DOWN group default 
    link/ether 02:42:06:d7:48:33 brd ff:ff:ff:ff:ff:ff
    inet 192.167.0.1/16 brd 192.167.255.255 scope global br-48dbf8e89480
       valid_lft forever preferred_lft forever

```

If you see a different IP address, update `../tendermint-pct-instrumentation/networks/local/localnode/config-template.toml` accordingly and rebuild the image.

### Testing service
You should also set the same host IP in `server.go` (in this repository), so it will bind to the correct interface:

```go
    // ...
	server, err := testlib.NewTestingServer(
		&config.Config{
			APIServerAddr: "192.167.0.1:7074",
            // ...
		},
        // ...
    )
```

### Running the tests
Start the testing server **before** the tendermint nodes:

```shell
go run ./server.go
```

In another terminal, start the tendermint nodes:

```shell
cd ../tenderint-pct-instrumentation
make localnet-start
```

Eventually you should see this line in the test output:
```json
{"level":"info","msg":"Testcase succeeded","service":"TestingServer","testcase":"RoundSkipWithPrevotes","time":"2022-05-20T11:11:06+02:00"}
```

You can then stop both the server and the nodes (with Ctrl-C).

By default the server runs the `RoundSkip` test. Other tests can be selected by uncommenting them in `server.go`. **Note that enabling multiple tests currently does not work**.

The expected results from the tests are as follows:
- **Passing**
    - RoundSkip
    - BlockVotes
    - CommitAfterRoundSkip
    - DifferentDecisions
    - ExpectUnlock
    - ExpectNoUnlock
    - LockedCommit
    - NilPrevotes
    - ProposalNilPrevote
    - ProposePrevote
    - QuorumPrevotes
    - NotNilDecide
    - LaggingReplica
    - HigherRound
    - CrashReplica
    - PrecommitsInvariant
- **Failing**
    - ExpectNewRound
    - Relocked
    - GarbledMessage
    - ForeverLaggingReplica
- **Flaky (fails sometimes)**
    - QuorumPrecommits

## Common issues
### Permission denied on second run
If you do not use rootless docker, running the nodes may create files owned by `root`, and on a second run you will see errors such as:

```
rm: cannot remove '/home/daan/workspace/tendermint-pct-instrumentation/build/node0/data/cs.wal': Permission denied
```

As a workaround, you can make this change to `../tendermint-pct-instrumentation/Makefile`:

```diff
    # Stop testnet
    localnet-stop:
    docker-compose down
-   rm -rf $(BUILDDIR)/node*
+   docker run --rm -v $(BUILDDIR):/tendermint alpine rm -rf /tendermint/node0 /tendermint/node1 /tendermint/node2 /tendermint/node3
    .PHONY: localnet-stop
```