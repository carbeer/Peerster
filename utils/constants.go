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

func GetSearchRequestTimeout() time.Duration {
	return 2 * time.Second
}

func GetAntiEntropyFrequency() time.Duration {
	return time.Second
}

func GetCachingDurationMS() time.Duration {
	return time.Millisecond * 500
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

func GetDefaultBudget() uint64 {
	return 2
}

func GetMaximumBudget() uint64 {
	return 32
}

func GetMinimumThreshold() uint32 {
	return 2
}

func GetTxPulishHopLimit() int {
	return 10
}

func GetBlockPublishHopLimit() int {
	return 20
}

func GetNumberOfLeadZeroes() int {
	return 16
}

func GetDownloadFolder() string {
	return "_Downloads"
}

func GetSharedFolder() string {
	return "_SharedFiles"
}
