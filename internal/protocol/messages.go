package protocol

import (
	"fmt"
	"strconv"
)

type Message []string

type Block []Message

var OKMessage = Message{"OK"}

func ErrorMessage(f string, args ...interface{}) Message {
	return Message{"ERROR", fmt.Sprintf(f, args...)}
}

func PathNameMessage(pathName string) Message {
	return Message{"PATHNAME", pathName}
}

var BoolTrueMessage = Message{"BOOL", "TRUE"}
var BoolFalseMessage = Message{"BOOL", "FALSE"}

func BoolMessage(v bool) Message {
	if v {
		return BoolTrueMessage
	} else {
		return BoolFalseMessage
	}
}

func HandshakeResponseMessage(version int, builder string, flags int) Message {
	return Message{
		"HELLO",
		strconv.FormatInt(int64(version), 10),
		builder,
		strconv.FormatInt(int64(flags), 10),
	}
}
