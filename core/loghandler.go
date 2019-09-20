package core

/*
#include <ctanker.h>

void tanker_log_handler_proxy(tanker_log_record_t *record);

static void _tanker_set_log_handler()
{
	tanker_set_log_handler((void(*)(tanker_log_record_t const*))(&tanker_log_handler_proxy));
}

*/
import "C"

type LogLevel rune
type LogRecord struct {
	Category string
	Level    LogLevel
	File     string
	Line     uint
	Message  string
}
type LogHandler func(LogRecord)

var currentLogHandler LogHandler = nil

func convertLogLevel(level C.uint) LogLevel {
	var logLevel LogLevel = 0
	switch level {
	case C.TANKER_LOG_DEBUG:
		logLevel = 'D'
	case C.TANKER_LOG_INFO:
		logLevel = 'I'
	case C.TANKER_LOG_WARNING:
		logLevel = 'W'
	case C.TANKER_LOG_ERROR:
		logLevel = 'E'
	}
	return logLevel
}

//export tanker_log_handler_proxy
func tanker_log_handler_proxy(crecord *C.tanker_log_record_t) {
	record := LogRecord{
		Category: C.GoString(crecord.category),
		Level:    convertLogLevel(crecord.level),
		File:     C.GoString(crecord.file),
		Line:     uint(crecord.line),
		Message:  C.GoString(crecord.message),
	}
	currentLogHandler(record)
}

//SetLogHandler set the loghandler of all tanker instances
func SetLogHandler(handler LogHandler) {
	currentLogHandler = handler

	C._tanker_set_log_handler()
}
