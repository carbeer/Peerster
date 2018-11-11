package utils

import (
	"time"
)

func GetHopLimitConstant() uint32 {
	return uint32(10)
}

func GetRumorMongeringTimeout() time.Duration {
	return time.Second
}

func GetDataRequestTimeout() time.Duration {
	return 5 * time.Second
}

func GetAntiEntropyFrequency() time.Duration {
	return time.Second
}

func GetClientIp() string {
	return "127.0.0.1"
}

func GetUIPort() string {
	return "8080"
}

func GetChunkSize() int {
	return 8192
}

func GetMsgBuffer() int {
	return 100
}
