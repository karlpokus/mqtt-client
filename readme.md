# mqtt-client
An mqtt client in go supporting [mqtt v3.1.1](http://docs.oasis-open.org/mqtt/mqtt/v3.1.1/os/mqtt-v3.1.1-os.html#_Toc398718086). Mostly for learning bit operations, not for production use.

# design
Composition over inheritance. Sending packets on a stream.

# notes
- Use `printf "%d\n" <hex>` and `printf "%x\n" <int>` to convert between integer and hex
- Use `<-func()` to indicate waiting
- `len(chan)` only shows unread items, not pending writers
- `SIGINT` takes a while since `DISCONNECT` is also time sharing the stream

# usage
````bash
$ make test
$ make run
````

# todos
- [x] use https://github.com/karlpokus/mqtt-parser for monitoring connection
- [x] connect
- [x] connack
- [x] pingreq
- [x] pingresp
- [x] disconnect
- [x] tests
- [ ] pub
- [x] parse connack return codes
- [x] verify pingresp after sending pingreq
- [x] handle conn errors https://github.com/karlpokus/broker/blob/master/pkg/broker/broker.go
- [x] try bytes.Buffer and buf.WriteTo(conn)
- [ ] close connection on disconnect
- [x] rw wrapper that log op codes
- [x] time share rw
- [ ] try buffered rwc
- [ ] stream.Fake() ?
- [x] rw error handling options: pass fatal or return err
- [x] move opFuncs from packet to stream
