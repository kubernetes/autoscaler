/*
 * ecs
 *
 * ECS Open API
 *
 */

package model

import (
	"encoding/json"

	"strings"
)

//
type JobEntities struct {
	// 每个子任务的执行信息。
	SubJobs *[]SubJob `json:"sub_jobs,omitempty"`
	// 子任务数量。
	SubJobsTotal *int32 `json:"sub_jobs_total,omitempty"`
}

func (o JobEntities) String() string {
	data, _ := json.Marshal(o)
	return strings.Join([]string{"JobEntities", string(data)}, " ")
}
