package main

import "sync"

var Handlers = map[string]func([]Value) Value{
	"PING":    ping,
	"SET":     set,
	"GET":     get,
	"HSET":    hset,
	"HGET":    hget,
	"HGETALL": hgetall,
}

func ping(values []Value) Value {
	if len(values) == 0 {
		return Value{typ: "string", str: "PONG"}
	}

	return Value{typ: "string", str: values[0].bulk}
}

var SETs = map[string]string{}
var SETsMu = sync.RWMutex{}

func set(values []Value) Value {
	if len(values) != 2 {
		return Value{typ: "error", str: "ERR wrong number of arguments"}
	}

	key := values[0].bulk
	val := values[1].bulk

	SETsMu.Lock()
	SETs[key] = val
	SETsMu.Unlock()

	return Value{typ: "string", str: "OK"}
}

func get(values []Value) Value {
	if len(values) != 1 {
		return Value{typ: "error", str: "ERR wrong number of arguments"}
	}

	key := values[0].bulk

	SETsMu.Lock()
	val, ok := SETs[key]
	SETsMu.Unlock()

	if !ok {
		return Value{typ: "null"}
	}

	return Value{typ: "bulk", bulk: val}
}

var HSETs = map[string]map[string]string{}
var HSETsMu = sync.RWMutex{}

func hset(values []Value) Value {
	if len(values) != 3 {
		return Value{typ: "error", str: "ERR wrong number of arguments"}
	}

	hash := values[0].bulk
	key := values[1].bulk
	val := values[2].bulk

	HSETsMu.Lock()
	if _, ok := HSETs[hash]; !ok {
		HSETs[hash] = map[string]string{}
	}

	HSETs[hash][key] = val
	HSETsMu.Unlock()

	return Value{typ: "string", str: "OK"}
}

func hget(values []Value) Value {
	if len(values) != 2 {
		return Value{typ: "error", str: "ERR wrong number of arguments"}
	}

	hash := values[0].bulk
	key := values[1].bulk

	HSETsMu.Lock()
	val, ok := HSETs[hash][key]
	HSETsMu.Unlock()

	if !ok {
		return Value{typ: "null"}
	}

	return Value{typ: "bulk", bulk: val}
}

func hgetall(values []Value) Value {
	if len(values) != 1 {
		return Value{typ: "error", str: "ERR wrong number of arguments"}
	}

	hash := values[0].bulk

	HSETsMu.Lock()
	val, ok := HSETs[hash]
	HSETsMu.Unlock()

	if !ok {
		return Value{typ: "null"}
	}

	vals := []Value{}
	for k, v := range val {
		vals = append(vals, Value{typ: "bulk", bulk: k})
		vals = append(vals, Value{typ: "bulk", bulk: v})
	}

	return Value{typ: "array", arr: vals}
}
