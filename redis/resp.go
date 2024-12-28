package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

const (
	STRING  = '+'
	ERROR   = '-'
	INTEGER = ':'
	BULK    = '$'
	ARRAY   = '*'
	CR      = '\r'
	NEWLINE = '\n'
)

type Value struct {
	typ  string
	str  string
	num  int
	bulk string
	arr  []Value
}

type Writer struct {
	writer io.Writer
}

func NewWriter(wr io.Writer) *Writer {
	return &Writer{writer: wr}
}

func (w *Writer) Write(v Value) error {
	_, err := w.writer.Write(v.Marshal())
	if err != nil {
		return err
	}

	return nil
}

type Resp struct {
	reader *bufio.Reader
}

func NewResp(rd io.Reader) *Resp {
	return &Resp{reader: bufio.NewReader(rd)}
}

func (r *Resp) readLine() (line []byte, n int, err error) {
	for {
		b, err := r.reader.ReadByte()
		if err != nil {
			return nil, 0, err
		}

		n += 1
		line = append(line, b)
		if len(line) >= 2 && line[len(line)-2] == CR {
			break
		}
	}

	return line[:len(line)-2], n, nil
}

func (r *Resp) readInteger() (x int, n int, err error) {
	line, n, err := r.readLine()
	if err != nil {
		return 0, 0, err
	}

	i64, err := strconv.ParseInt(string(line), 10, 64)
	if err != nil {
		return 0, 0, err
	}

	return int(i64), n, nil
}

func (r *Resp) Read() (Value, error) {
	_type, err := r.reader.ReadByte()
	if err != nil {
		return Value{}, err
	}

	switch _type {
	case ARRAY:
		return r.readArray()
	case BULK:
		return r.readBulk()
	default:
		fmt.Printf("Unknown type: %v", string(_type))
		return Value{}, nil
	}
}

func (r *Resp) readArray() (Value, error) {
	v := Value{typ: "array"}

	length, _, err := r.readInteger()
	if err != nil {
		return v, err
	}

	v.arr = make([]Value, length)
	for i := 0; i < length; i++ {
		val, err := r.Read()
		if err != nil {
			return v, err
		}
		v.arr[i] = val
	}

	return v, nil
}

func (r *Resp) readBulk() (Value, error) {
	v := Value{typ: "bulk"}

	length, _, err := r.readInteger()
	if err != nil {
		return v, err
	}

	bulk := make([]byte, length)
	_, err = r.reader.Read(bulk)
	if err != nil {
		return v, err
	}

	v.bulk = string(bulk)

	r.readLine()

	return v, nil
}

func (v Value) Marshal() []byte {
	switch v.typ {
	case "array":
		return v.marshalArray()
	case "bulk":
		return v.marshalBulk()
	case "string":
		return v.marshalString()
	case "error":
		return v.marshallError()
	case "null":
		return v.marshallNull()
	default:
		return []byte{}
	}
}

func (v Value) marshalArray() []byte {
	length := len(v.arr)
	var bytes []byte

	bytes = append(bytes, ARRAY)
	bytes = append(bytes, strconv.Itoa(length)...)
	bytes = append(bytes, CR, NEWLINE)

	for i := 0; i < length; i++ {
		bytes = append(bytes, v.arr[i].Marshal()...)
	}

	return bytes
}

func (v Value) marshalBulk() []byte {
	length := len(v.bulk)
	var bytes []byte

	bytes = append(bytes, BULK)
	bytes = append(bytes, strconv.Itoa(length)...)
	bytes = append(bytes, CR, NEWLINE)
	bytes = append(bytes, v.bulk...)
	bytes = append(bytes, CR, NEWLINE)

	return bytes
}

func (v Value) marshalString() []byte {
	var bytes []byte
	bytes = append(bytes, STRING)
	bytes = append(bytes, v.str...)
	bytes = append(bytes, CR, NEWLINE)

	return bytes
}

func (v Value) marshallError() []byte {
	var bytes []byte
	bytes = append(bytes, ERROR)
	bytes = append(bytes, v.str...)
	bytes = append(bytes, CR, NEWLINE)

	return bytes
}

func (v Value) marshallNull() []byte {
	return []byte("$-1\r\n")
}
