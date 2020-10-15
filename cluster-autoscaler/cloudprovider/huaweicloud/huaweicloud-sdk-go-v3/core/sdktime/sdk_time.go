package sdktime

import (
	"strings"
	"time"
)

type SdkTime time.Time

func (t *SdkTime) UnmarshalJSON(data []byte) error {
	tmp := strings.Trim(string(data[:]), "\"")
	now, err := time.ParseInLocation(`2006-01-02T15:04:05Z`, tmp, time.UTC)
	if err != nil {
		now, err = time.ParseInLocation(`2006-01-02T15:04:05`, tmp, time.UTC)
		if err != nil {
			now, err = time.ParseInLocation(`2006-01-02 15:04:05`, tmp, time.UTC)
			if err != nil {
				return err
			}
		}
	}
	*t = SdkTime(now)
	return nil
}

func (t SdkTime) String() string {
	return time.Time(t).Format(`2006-01-02T15:04:05Z`)
}
