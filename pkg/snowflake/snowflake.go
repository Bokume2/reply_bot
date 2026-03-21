package snowflake

import "time"

func TimeToSnowflake(t time.Time, inc uint16) (sf uint64) {
	sf = uint64(t.UnixMilli() & 0xFFFFFFFFFFFF)
	sf = sf << 16
	sf += uint64(inc)
	return
}
