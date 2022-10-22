# mqtt-client
An mqtt client in go supporting [mqtt v3.1.1](http://docs.oasis-open.org/mqtt/mqtt/v3.1.1/os/mqtt-v3.1.1-os.html#_Toc398718086). Mostly for learning bit operations, not for production use.

# notes
- Use `printf "%d\n" <hex>` and `printf "%x\n" <int>` to convert between integer and hex

# usage
````bash
$ go run .
````

# todos
- [x] use https://github.com/karlpokus/mqtt-parser for monitoring connection
- [x] connect
- [x] connack
- [x] pingreq
- [x] pingresp
- [x] disconnect
- [ ] tests
- [ ] pub
- [ ] parse connack return codes
- [ ] verify pingresp after sending pingreq
- [ ] handle conn errors https://github.com/karlpokus/broker/blob/master/pkg/broker/broker.go
- [ ] try bytes.Buffer and buf.WriteTo(conn)
