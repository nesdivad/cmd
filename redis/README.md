# Redis server

- Building a redis server by following this guide: https://www.build-redis-from-scratch.dev/

## Redis format

The command `SET key value` corresponds to:
- `*3\r\n$3\r\nset\r\n$3\r\nkey\r\n$5\r\nvalue\r\n`