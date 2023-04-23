package utils

import (
  "math"
  "time"
)

func NotTimeUTC() time.Time {
  return time.Now().UTC()
}

func TimeToUTC(t time.Time) time.Time {
  return t.UTC()
}

func TimestampToTimeUTC(ts int64) time.Time {
  const (
    thresholdYears = 100 // for correct
  )
  // milliseconds epoch
  t := time.UnixMilli(ts).UTC()
  if math.Abs(float64(time.Now().UTC().Year()-t.Year())) > thresholdYears {
    // seconds epoch
    t = time.Unix(ts, 0).UTC()
  }
  return t
}

func NowTimestampUTC() int64 {
  return time.Now().Unix()
}
