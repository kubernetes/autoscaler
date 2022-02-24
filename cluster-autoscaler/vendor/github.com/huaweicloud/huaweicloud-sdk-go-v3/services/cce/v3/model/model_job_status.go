package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

//
type JobStatus struct {
	// 任务的状态，有如下四种状态：  - JobPhaseInitializing JobPhase = \"Initializing\" - JobPhaseRunning JobPhase = \"Running\" - JobPhaseFailed JobPhase = \"Failed\" - JobPhaseSuccess JobPhase = \"Success\"

	Phase *string `json:"phase,omitempty"`
	// 任务变为当前状态的原因

	Reason *string `json:"reason,omitempty"`
}

func (o JobStatus) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "JobStatus struct{}"
	}

	return strings.Join([]string{"JobStatus", string(data)}, " ")
}
