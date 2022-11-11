# mqtt-client
An mqtt client in go supporting [mqtt v3.1.1](http://docs.oasis-open.org/mqtt/mqtt/v3.1.1/os/mqtt-v3.1.1-os.html#_Toc398718086). Mostly for learning bit operations, not for production use.

# mqtt 3.1.1 feature support
- [x] connect
- [x] ping
- [x] subscribe
- [x] disconnect
- [ ] publish
- [ ] QoS
- [ ] session

# design goals
- Composition over inheritance
- Simplicity: just packets and a stream
- Idiomatic go
- Decent test coverage

# notes
- Use `printf "%d\n" <hex>` and `printf "%x\n" <int>` to convert between integer and hex
- Use `<-func()` to indicate waiting
- `len(chan)` only shows unread items, not pending writers
- `SIGINT` takes a while since `DISCONNECT` is also time sharing the stream
- Use `Set*Deadline` to cancel `net.Conn.Read`
- Use `context` instead of closing a channel to abort reading a channel if you have multiple writers
- Empty channels do not need to be closed. They will be garbage collected.
- Stopping a `time.Timer` does not close `time.Timer.C` so stopping a timer while we're reading from it in another goroutine will leave the channel open. Will this leak that goroutine? Or will gb reap it?

# usage
````bash
$ make test
$ make run
````

# todos
- [x] use https://github.com/karlpokus/mqtt-parser for monitoring connection
- [x] tests
- [x] parse connack return codes
- [x] verify pingresp after sending pingreq
- [x] handle conn errors https://github.com/karlpokus/broker/blob/master/pkg/broker/broker.go
- [x] try bytes.Buffer and buf.WriteTo(conn)
- [ ] close connection on disconnect
- [x] rw wrapper that log op codes
- [x] time share rw
- [x] try buffered rwc - no
- [x] stream.Fake
- [x] rw error handling options: pass fatal or return err
- [x] move opFuncs from packet to stream
- [ ] type op execution timeout
- [x] implement read timeout for io.ReadWriter
- [x] client ui
- [ ] client config
- [ ] client topic router
- [ ] prefix log records
- [x] stop ops listener before exit
- [x] pass error in Response
- [x] ack pending queue
- [ ] maybe rename type op to tx
- [ ] good comments on packet bytes
- [x] expose Response.Message and Topic
- [ ] *ack should include its name in timeout error
- [x] not before SUBACK popped should we notify client
- [x] ack name
- [ ] reconnect option
- [x] make op funcs cancelable
- [ ] slow reader to test read timeout
- [x] generate packet id
