package protocol

type Message []string

type Block []Message

var OKMessage = Message{"OK"}

func ErrorMessage(message string) Message {
	return Message{"ERROR", message}
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
