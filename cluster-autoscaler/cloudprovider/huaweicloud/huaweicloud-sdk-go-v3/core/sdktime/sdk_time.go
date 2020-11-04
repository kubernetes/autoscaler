package sdktime

import (
	"fmt"
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

func (t SdkTime) MarshalJSON() ([]byte, error) {
	rs := []byte(fmt.Sprintf(`"%s"`, t.String()))
	return rs, nil
}

func (t SdkTime) String() string {
	// return time.Time(t).Format(`2006-01-02T15:04:05Z`)
	// temp solution for: https://github.com/huaweicloud/huaweicloud-sdk-go-v3/issues/8
	return time.Time(t).Format(`2006-01-02T15:04Z`)
}
