package mrscaler

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/spotinst/spotinst-sdk-go/spotinst"
	"github.com/spotinst/spotinst-sdk-go/spotinst/client"
	"github.com/spotinst/spotinst-sdk-go/spotinst/util/jsonutil"
	"github.com/spotinst/spotinst-sdk-go/spotinst/util/uritemplates"
)

// A InstanceGroupType represents the type of an instance group.
type InstanceGroupType int

const (
	// InstanceGroupTypeMaster represents the master instance group type.
	InstanceGroupTypeMaster InstanceGroupType = iota

	// InstanceGroupTypeCore represents the core instance group type.
	InstanceGroupTypeCore

	// InstanceGroupTypeTask represents the task instance group type.
	InstanceGroupTypeTask
)

var InstanceGroupType_name = map[InstanceGroupType]string{
	InstanceGroupTypeMaster: "master",
	InstanceGroupTypeCore:   "core",
	InstanceGroupTypeTask:   "task",
}

var InstanceGroupType_value = map[string]InstanceGroupType{
	"master": InstanceGroupTypeMaster,
	"core":   InstanceGroupTypeCore,
	"task":   InstanceGroupTypeTask,
}

func (p InstanceGroupType) String() string {
	return InstanceGroupType_name[p]
}

type Scaler struct {
	ID          *string   `json:"id,omitempty"`
	Name        *string   `json:"name,omitempty"`
	Description *string   `json:"description,omitempty"`
	Region      *string   `json:"region,omitempty"`
	Strategy    *Strategy `json:"strategy,omitempty"`
	Compute     *Compute  `json:"compute,omitempty"`
	Scaling     *Scaling  `json:"scaling,omitempty"`
	CoreScaling *Scaling  `json:"coreScaling,omitempty"`

	// forceSendFields is a list of field names (e.g. "Keys") to
	// unconditionally include in API requests. By default, fields with
	// empty values are omitted from API requests. However, any non-pointer,
	// non-interface field appearing in ForceSendFields will be sent to the
	// server regardless of whether the field is empty or not. This may be
	// used to include empty fields in Patch requests.
	forceSendFields []string `json:"-"`

	// nullFields is a list of field names (e.g. "Keys") to include in API
	// requests with the JSON null value. By default, fields with empty
	// values are omitted from API requests. However, any field with an
	// empty value appearing in NullFields will be sent to the server as
	// null. It is an error if a field in this list has a non-empty value.
	// This may be used to include null fields in Patch requests.
	nullFields []string `json:"-"`
}

type Strategy struct {
	Cloning  *Cloning  `json:"cloning,omitempty"`
	Wrapping *Wrapping `json:"wrapping,omitempty"`

	forceSendFields []string `json:"-"`
	nullFields      []string `json:"-"`
}

type Cloning struct {
	OriginClusterID *string `json:"originClusterId,omitempty"`

	forceSendFields []string `json:"-"`
	nullFields      []string `json:"-"`
}

type Wrapping struct {
	SourceClusterID *string `json:"sourceClusterId,omitempty"`

	forceSendFields []string `json:"-"`
	nullFields      []string `json:"-"`
}

type Compute struct {
	AvailabilityZones []*AvailabilityZone `json:"availabilityZones,omitempty"`
	Tags              []*Tag              `json:"tags,omitempty"`
	InstanceGroups    *InstanceGroups     `json:"instanceGroups,omitempty"`
	Configurations    *Configurations     `json:"configurations,omitempty"`

	forceSendFields []string `json:"-"`
	nullFields      []string `json:"-"`
}

type AvailabilityZone struct {
	Name     *string `json:"name,omitempty"`
	SubnetID *string `json:"subnetId,omitempty"`

	forceSendFields []string `json:"-"`
	nullFields      []string `json:"-"`
}

type Tag struct {
	Key   *string `json:"tagKey,omitempty"`
	Value *string `json:"tagValue,omitempty"`

	forceSendFields []string `json:"-"`
	nullFields      []string `json:"-"`
}

type InstanceGroups struct {
	MasterGroup *InstanceGroup `json:"masterGroup,omitempty"`
	CoreGroup   *InstanceGroup `json:"coreGroup,omitempty"`
	TaskGroup   *InstanceGroup `json:"taskGroup,omitempty"`

	forceSendFields []string `json:"-"`
	nullFields      []string `json:"-"`
}

type InstanceGroup struct {
	InstanceTypes    []string               `json:"instanceTypes,omitempty"`
	Target           *int                   `json:"target,omitempty"`
	Capacity         *InstanceGroupCapacity `json:"capacity,omitempty"`
	LifeCycle        *string                `json:"lifeCycle,omitempty"`
	EBSConfiguration *EBSConfiguration      `json:"ebsConfiguration,omitempty"`

	forceSendFields []string `json:"-"`
	nullFields      []string `json:"-"`
}

type InstanceGroupCapacity struct {
	Target  *int `json:"target,omitempty"`
	Minimum *int `json:"minimum,omitempty"`
	Maximum *int `json:"maximum,omitempty"`

	forceSendFields []string `json:"-"`
	nullFields      []string `json:"-"`
}

type EBSConfiguration struct {
	Optimized          *bool                `json:"ebsOptimized,omitempty"`
	BlockDeviceConfigs []*BlockDeviceConfig `json:"ebsBlockDeviceConfigs,omitempty"`

	forceSendFields []string `json:"-"`
	nullFields      []string `json:"-"`
}

type BlockDeviceConfig struct {
	VolumesPerInstance  *int                 `json:"volumesPerInstance,omitempty"`
	VolumeSpecification *VolumeSpecification `json:"volumeSpecification,omitempty"`

	forceSendFields []string `json:"-"`
	nullFields      []string `json:"-"`
}

type VolumeSpecification struct {
	VolumeType *string `json:"volumeType,omitempty"`
	SizeInGB   *int    `json:"sizeInGB,omitempty"`
	IOPS       *int    `json:"iops,omitempty"`

	forceSendFields []string `json:"-"`
	nullFields      []string `json:"-"`
}

type Scaling struct {
	Up   []*ScalingPolicy `json:"up,omitempty"`
	Down []*ScalingPolicy `json:"down,omitempty"`

	forceSendFields []string `json:"-"`
	nullFields      []string `json:"-"`
}

type ScalingPolicy struct {
	PolicyName        *string      `json:"policyName,omitempty"`
	Namespace         *string      `json:"namespace,omitempty"`
	MetricName        *string      `json:"metricName,omitempty"`
	Dimensions        []*Dimension `json:"dimensions,omitempty"`
	Statistic         *string      `json:"statistic,omitempty"`
	Unit              *string      `json:"unit,omitempty"`
	Threshold         *float64     `json:"threshold,omitempty"`
	Period            *int         `json:"period,omitempty"`
	EvaluationPeriods *int         `json:"evaluationPeriods,omitempty"`
	Cooldown          *int         `json:"cooldown,omitempty"`
	Action            *Action      `json:"action,omitempty"`
	Operator          *string      `json:"operator,omitempty"`

	forceSendFields []string `json:"-"`
	nullFields      []string `json:"-"`
}

type Action struct {
	Type              *string `json:"type,omitempty"`
	Adjustment        *string `json:"adjustment,omitempty"`
	MinTargetCapacity *string `json:"minTargetCapacity,omitempty"`
	MaxTargetCapacity *string `json:"maxTargetCapacity,omitempty"`
	Target            *string `json:"target,omitempty"`
	Minimum           *string `json:"minimum,omitempty"`
	Maximum           *string `json:"maximum,omitempty"`

	forceSendFields []string `json:"-"`
	nullFields      []string `json:"-"`
}

type Dimension struct {
	Name  *string `json:"name,omitempty"`
	Value *string `json:"value,omitempty"`

	forceSendFields []string `json:"-"`
	nullFields      []string `json:"-"`
}

type Configurations struct {
	File *ConfigurationFile `json:"file,omitempty"`

	forceSendFields []string `json:"-"`
	nullFields      []string `json:"-"`
}

type ConfigurationFile struct {
	Bucket *string `json:"bucket,omitempty"`
	Key    *string `json:"key,omitempty"`

	forceSendFields []string `json:"-"`
	nullFields      []string `json:"-"`
}

type ListScalersInput struct{}

type ListScalersOutput struct {
	Scalers []*Scaler `json:"mrScalers,omitempty"`
}

type CreateScalerInput struct {
	Scaler *Scaler `json:"mrScaler,omitempty"`
}

type CreateScalerOutput struct {
	Scaler *Scaler `json:"mrScaler,omitempty"`
}

type ReadScalerInput struct {
	ScalerID *string `json:"mrScalerId,omitempty"`
}

type ReadScalerOutput struct {
	Scaler *Scaler `json:"mrScaler,omitempty"`
}

type UpdateScalerInput struct {
	Scaler *Scaler `json:"mrScaler,omitempty"`
}

type UpdateScalerOutput struct {
	Scaler *Scaler `json:"mrScaler,omitempty"`
}

type DeleteScalerInput struct {
	ScalerID *string `json:"mrScalerId,omitempty"`
}

type DeleteScalerOutput struct{}

type StatusScalerInput struct {
	ScalerID *string `json:"mrScalerId,omitempty"`
}

type StatusScalerOutput struct {
	Instances []*Instance `json:"instances,omitempty"`
}

type Instance struct {
	ID               *string    `json:"instanceId,omitempty"`
	SpotRequestID    *string    `json:"spotInstanceRequestId,omitempty"`
	InstanceType     *string    `json:"instanceType,omitempty"`
	Status           *string    `json:"status,omitempty"`
	Product          *string    `json:"product,omitempty"`
	AvailabilityZone *string    `json:"availabilityZone,omitempty"`
	CreatedAt        *time.Time `json:"createdAt,omitempty"`
}

func scalerFromJSON(in []byte) (*Scaler, error) {
	b := new(Scaler)
	if err := json.Unmarshal(in, b); err != nil {
		return nil, err
	}
	return b, nil
}

func scalersFromJSON(in []byte) ([]*Scaler, error) {
	var rw client.Response
	if err := json.Unmarshal(in, &rw); err != nil {
		return nil, err
	}
	out := make([]*Scaler, len(rw.Response.Items))
	if len(out) == 0 {
		return out, nil
	}
	for i, rb := range rw.Response.Items {
		b, err := scalerFromJSON(rb)
		if err != nil {
			return nil, err
		}
		out[i] = b
	}
	return out, nil
}

func scalersFromHttpResponse(resp *http.Response) ([]*Scaler, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return scalersFromJSON(body)
}

//region Scaler

func (o *Scaler) MarshalJSON() ([]byte, error) {
	type noMethod Scaler
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Scaler) SetId(v *string) *Scaler {
	if o.ID = v; v == nil {
		o.nullFields = append(o.nullFields, "ID")
	}
	return o
}

func (o *Scaler) SetName(v *string) *Scaler {
	if o.Name = v; v == nil {
		o.nullFields = append(o.nullFields, "Name")
	}
	return o
}

func (o *Scaler) SetDescription(v *string) *Scaler {
	if o.Description = v; v == nil {
		o.nullFields = append(o.nullFields, "Description")
	}
	return o
}

func (o *Scaler) SetRegion(v *string) *Scaler {
	if o.Region = v; v == nil {
		o.nullFields = append(o.nullFields, "Region")
	}
	return o
}

func (o *Scaler) SetStrategy(v *Strategy) *Scaler {
	if o.Strategy = v; v == nil {
		o.nullFields = append(o.nullFields, "Strategy")
	}
	return o
}

func (o *Scaler) SetCompute(v *Compute) *Scaler {
	if o.Compute = v; v == nil {
		o.nullFields = append(o.nullFields, "Compute")
	}
	return o
}

func (o *Scaler) SetScaling(v *Scaling) *Scaler {
	if o.Scaling = v; v == nil {
		o.nullFields = append(o.nullFields, "Scaling")
	}
	return o
}

func (o *Scaler) SetCoreScaling(v *Scaling) *Scaler {
	if o.CoreScaling = v; v == nil {
		o.nullFields = append(o.nullFields, "CoreScaling")
	}
	return o
}

//endregion

//region Strategy

func (o *Strategy) MarshalJSON() ([]byte, error) {
	type noMethod Strategy
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Strategy) SetCloning(v *Cloning) *Strategy {
	if o.Cloning = v; v == nil {
		o.nullFields = append(o.nullFields, "Cloning")
	}
	return o
}

func (o *Strategy) SetWrapping(v *Wrapping) *Strategy {
	if o.Wrapping = v; v == nil {
		o.nullFields = append(o.nullFields, "Wrapping")
	}
	return o
}

//endregion

//region Cloning

func (o *Cloning) MarshalJSON() ([]byte, error) {
	type noMethod Cloning
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Cloning) SetOriginClusterId(v *string) *Cloning {
	if o.OriginClusterID = v; v == nil {
		o.nullFields = append(o.nullFields, "OriginClusterID")
	}
	return o
}

//endregion

//region Wrapping

func (o *Wrapping) MarshalJSON() ([]byte, error) {
	type noMethod Wrapping
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Wrapping) SetSourceClusterId(v *string) *Wrapping {
	if o.SourceClusterID = v; v == nil {
		o.nullFields = append(o.nullFields, "SourceClusterID")
	}
	return o
}

//endregion

//region Compute

func (o *Compute) MarshalJSON() ([]byte, error) {
	type noMethod Compute
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Compute) SetAvailabilityZones(v []*AvailabilityZone) *Compute {
	if o.AvailabilityZones = v; v == nil {
		o.nullFields = append(o.nullFields, "AvailabilityZones")
	}
	return o
}

func (o *Compute) SetTags(v []*Tag) *Compute {
	if o.Tags = v; v == nil {
		o.nullFields = append(o.nullFields, "Tags")
	}
	return o
}

func (o *Compute) SetInstanceGroups(v *InstanceGroups) *Compute {
	if o.InstanceGroups = v; v == nil {
		o.nullFields = append(o.nullFields, "InstanceGroups")
	}
	return o
}

func (o *Compute) SetConfigurations(v *Configurations) *Compute {
	if o.Configurations = v; v == nil {
		o.nullFields = append(o.nullFields, "Configurations")
	}
	return o
}

//endregion

//region AvailabilityZone

func (o *AvailabilityZone) MarshalJSON() ([]byte, error) {
	type noMethod AvailabilityZone
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *AvailabilityZone) SetName(v *string) *AvailabilityZone {
	if o.Name = v; v == nil {
		o.nullFields = append(o.nullFields, "Name")
	}
	return o
}

func (o *AvailabilityZone) SetSubnetId(v *string) *AvailabilityZone {
	if o.SubnetID = v; v == nil {
		o.nullFields = append(o.nullFields, "SubnetID")
	}
	return o
}

//endregion

//region Tag

func (o *Tag) MarshalJSON() ([]byte, error) {
	type noMethod Tag
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Tag) SetKey(v *string) *Tag {
	if o.Key = v; v == nil {
		o.nullFields = append(o.nullFields, "Key")
	}
	return o
}

func (o *Tag) SetValue(v *string) *Tag {
	if o.Value = v; v == nil {
		o.nullFields = append(o.nullFields, "Value")
	}
	return o
}

//endregion

//region InstanceGroups

func (o *InstanceGroups) MarshalJSON() ([]byte, error) {
	type noMethod InstanceGroups
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *InstanceGroups) SetMasterGroup(v *InstanceGroup) *InstanceGroups {
	if o.MasterGroup = v; v == nil {
		o.nullFields = append(o.nullFields, "MasterGroup")
	}
	return o
}

func (o *InstanceGroups) SetCoreGroup(v *InstanceGroup) *InstanceGroups {
	if o.CoreGroup = v; v == nil {
		o.nullFields = append(o.nullFields, "CoreGroup")
	}
	return o
}

func (o *InstanceGroups) SetTaskGroup(v *InstanceGroup) *InstanceGroups {
	if o.TaskGroup = v; v == nil {
		o.nullFields = append(o.nullFields, "TaskGroup")
	}
	return o
}

//endregion

//region InstanceGroup

func (o *InstanceGroup) MarshalJSON() ([]byte, error) {
	type noMethod InstanceGroup
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *InstanceGroup) SetInstanceTypes(v []string) *InstanceGroup {
	if o.InstanceTypes = v; v == nil {
		o.nullFields = append(o.nullFields, "InstanceTypes")
	}
	return o
}

func (o *InstanceGroup) SetTarget(v *int) *InstanceGroup {
	if o.Target = v; v == nil {
		o.nullFields = append(o.nullFields, "Target")
	}
	return o
}

func (o *InstanceGroup) SetCapacity(v *InstanceGroupCapacity) *InstanceGroup {
	if o.Capacity = v; v == nil {
		o.nullFields = append(o.nullFields, "Capacity")
	}
	return o
}

func (o *InstanceGroup) SetLifeCycle(v *string) *InstanceGroup {
	if o.LifeCycle = v; v == nil {
		o.nullFields = append(o.nullFields, "LifeCycle")
	}
	return o
}

func (o *InstanceGroup) SetEBSConfiguration(v *EBSConfiguration) *InstanceGroup {
	if o.EBSConfiguration = v; v == nil {
		o.nullFields = append(o.nullFields, "EBSConfiguration")
	}
	return o
}

//endregion

//region InstanceGroupCapacity
func (o *InstanceGroupCapacity) MarshalJSON() ([]byte, error) {
	type noMethod InstanceGroupCapacity
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *InstanceGroupCapacity) SetTarget(v *int) *InstanceGroupCapacity {
	if o.Target = v; v == nil {
		o.nullFields = append(o.nullFields, "Target")
	}
	return o
}

func (o *InstanceGroupCapacity) SetMinimum(v *int) *InstanceGroupCapacity {
	if o.Minimum = v; v == nil {
		o.nullFields = append(o.nullFields, "Minimum")
	}
	return o
}

func (o *InstanceGroupCapacity) SetMaximum(v *int) *InstanceGroupCapacity {
	if o.Maximum = v; v == nil {
		o.nullFields = append(o.nullFields, "Maximum")
	}
	return o
}

//endregion

//region EBSConfiguration
func (o *EBSConfiguration) MarshalJSON() ([]byte, error) {
	type noMethod EBSConfiguration
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *EBSConfiguration) SetOptimized(v *bool) *EBSConfiguration {
	if o.Optimized = v; v == nil {
		o.nullFields = append(o.nullFields, "Optimized")
	}
	return o
}

func (o *EBSConfiguration) SetBlockDeviceConfigs(v []*BlockDeviceConfig) *EBSConfiguration {
	if o.BlockDeviceConfigs = v; v == nil {
		o.nullFields = append(o.nullFields, "BlockDeviceConfigs")
	}
	return o
}

//endregion

//region BlockDeviceConfig
func (o *BlockDeviceConfig) MarshalJSON() ([]byte, error) {
	type noMethod BlockDeviceConfig
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *BlockDeviceConfig) SetVolumesPerInstance(v *int) *BlockDeviceConfig {
	if o.VolumesPerInstance = v; v == nil {
		o.nullFields = append(o.nullFields, "VolumesPerInstance")
	}
	return o
}

func (o *BlockDeviceConfig) SetVolumeSpecification(v *VolumeSpecification) *BlockDeviceConfig {
	if o.VolumeSpecification = v; v == nil {
		o.nullFields = append(o.nullFields, "VolumeSpecification")
	}
	return o
}

//endregion

//region VolumeSpecification
func (o *VolumeSpecification) MarshalJSON() ([]byte, error) {
	type noMethod VolumeSpecification
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *VolumeSpecification) SetVolumeType(v *string) *VolumeSpecification {
	if o.VolumeType = v; v == nil {
		o.nullFields = append(o.nullFields, "VolumeType")
	}
	return o
}

func (o *VolumeSpecification) SetSizeInGB(v *int) *VolumeSpecification {
	if o.SizeInGB = v; v == nil {
		o.nullFields = append(o.nullFields, "SizeInGB")
	}
	return o
}

func (o *VolumeSpecification) SetIOPS(v *int) *VolumeSpecification {
	if o.IOPS = v; v == nil {
		o.nullFields = append(o.nullFields, "IOPS")
	}
	return o
}

//endregion

//region Scaling

func (o *Scaling) MarshalJSON() ([]byte, error) {
	type noMethod Scaling
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Scaling) SetUp(v []*ScalingPolicy) *Scaling {
	if o.Up = v; v == nil {
		o.nullFields = append(o.nullFields, "Up")
	} else if len(o.Down) == 0 {
		o.forceSendFields = append(o.forceSendFields, "Up")
	}
	return o
}

func (o *Scaling) SetDown(v []*ScalingPolicy) *Scaling {
	if o.Down = v; v == nil {
		o.nullFields = append(o.nullFields, "Down")
	} else if len(o.Down) == 0 {
		o.forceSendFields = append(o.forceSendFields, "Down")
	}
	return o
}

//endregion

//region ScalingPolicy

func (o *ScalingPolicy) MarshalJSON() ([]byte, error) {
	type noMethod ScalingPolicy
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *ScalingPolicy) SetPolicyName(v *string) *ScalingPolicy {
	if o.PolicyName = v; v == nil {
		o.nullFields = append(o.nullFields, "PolicyName")
	}
	return o
}

func (o *ScalingPolicy) SetNamespace(v *string) *ScalingPolicy {
	if o.Namespace = v; v == nil {
		o.nullFields = append(o.nullFields, "Namespace")
	}
	return o
}

func (o *ScalingPolicy) SetMetricName(v *string) *ScalingPolicy {
	if o.MetricName = v; v == nil {
		o.nullFields = append(o.nullFields, "MetricName")
	}
	return o
}

func (o *ScalingPolicy) SetDimensions(v []*Dimension) *ScalingPolicy {
	if o.Dimensions = v; v == nil {
		o.nullFields = append(o.nullFields, "Dimensions")
	}
	return o
}

func (o *ScalingPolicy) SetStatistic(v *string) *ScalingPolicy {
	if o.Statistic = v; v == nil {
		o.nullFields = append(o.nullFields, "Statistic")
	}
	return o
}

func (o *ScalingPolicy) SetUnit(v *string) *ScalingPolicy {
	if o.Unit = v; v == nil {
		o.nullFields = append(o.nullFields, "Unit")
	}
	return o
}

func (o *ScalingPolicy) SetThreshold(v *float64) *ScalingPolicy {
	if o.Threshold = v; v == nil {
		o.nullFields = append(o.nullFields, "Threshold")
	}
	return o
}

func (o *ScalingPolicy) SetPeriod(v *int) *ScalingPolicy {
	if o.Period = v; v == nil {
		o.nullFields = append(o.nullFields, "Period")
	}
	return o
}

func (o *ScalingPolicy) SetEvaluationPeriods(v *int) *ScalingPolicy {
	if o.EvaluationPeriods = v; v == nil {
		o.nullFields = append(o.nullFields, "EvaluationPeriods")
	}
	return o
}

func (o *ScalingPolicy) SetCooldown(v *int) *ScalingPolicy {
	if o.Cooldown = v; v == nil {
		o.nullFields = append(o.nullFields, "Cooldown")
	}
	return o
}

func (o *ScalingPolicy) SetAction(v *Action) *ScalingPolicy {
	if o.Action = v; v == nil {
		o.nullFields = append(o.nullFields, "Action")
	}
	return o
}

func (o *ScalingPolicy) SetOperator(v *string) *ScalingPolicy {
	if o.Operator = v; v == nil {
		o.nullFields = append(o.nullFields, "Operator")
	}
	return o
}

//endregion

//region Action

func (o *Action) MarshalJSON() ([]byte, error) {
	type noMethod Action
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Action) SetType(v *string) *Action {
	if o.Type = v; v == nil {
		o.nullFields = append(o.nullFields, "Type")
	}
	return o
}

func (o *Action) SetAdjustment(v *string) *Action {
	if o.Adjustment = v; v == nil {
		o.nullFields = append(o.nullFields, "Adjustment")
	}
	return o
}

func (o *Action) SetMinTargetCapacity(v *string) *Action {
	if o.MinTargetCapacity = v; v == nil {
		o.nullFields = append(o.nullFields, "MinTargetCapacity")
	}
	return o
}

func (o *Action) SetMaxTargetCapacity(v *string) *Action {
	if o.MaxTargetCapacity = v; v == nil {
		o.nullFields = append(o.nullFields, "MaxTargetCapacity")
	}
	return o
}

func (o *Action) SetTarget(v *string) *Action {
	if o.Target = v; v == nil {
		o.nullFields = append(o.nullFields, "Target")
	}
	return o
}

func (o *Action) SetMinimum(v *string) *Action {
	if o.Minimum = v; v == nil {
		o.nullFields = append(o.nullFields, "Minimum")
	}
	return o
}

func (o *Action) SetMaximum(v *string) *Action {
	if o.Maximum = v; v == nil {
		o.nullFields = append(o.nullFields, "Maximum")
	}
	return o
}

//endregion

//region Dimension

func (o *Dimension) MarshalJSON() ([]byte, error) {
	type noMethod Dimension
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Dimension) SetName(v *string) *Dimension {
	if o.Name = v; v == nil {
		o.nullFields = append(o.nullFields, "Name")
	}
	return o
}

func (o *Dimension) SetValue(v *string) *Dimension {
	if o.Value = v; v == nil {
		o.nullFields = append(o.nullFields, "Value")
	}
	return o
}

//endregion

//region Configurations

func (o *Configurations) MarshalJSON() ([]byte, error) {
	type noMethod Configurations
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *Configurations) SetFile(v *ConfigurationFile) *Configurations {
	if o.File = v; v == nil {
		o.nullFields = append(o.nullFields, "File")
	}
	return o
}

//endregion

//region ConfigurationFile
func (o *ConfigurationFile) MarshalJSON() ([]byte, error) {
	type noMethod ConfigurationFile
	raw := noMethod(*o)
	return jsonutil.MarshalJSON(raw, o.forceSendFields, o.nullFields)
}

func (o *ConfigurationFile) SetBucket(v *string) *ConfigurationFile {
	if o.Bucket = v; v == nil {
		o.nullFields = append(o.nullFields, "Bucket")
	}
	return o
}

func (o *ConfigurationFile) SetKey(v *string) *ConfigurationFile {
	if o.Key = v; v == nil {
		o.nullFields = append(o.nullFields, "Key")
	}
	return o
}

//endregion

func (s *ServiceOp) List(ctx context.Context, input *ListScalersInput) (*ListScalersOutput, error) {
	r := client.NewRequest(http.MethodGet, "/aws/emr/mrScaler")
	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	gs, err := scalersFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	return &ListScalersOutput{Scalers: gs}, nil
}

func (s *ServiceOp) Create(ctx context.Context, input *CreateScalerInput) (*CreateScalerOutput, error) {
	r := client.NewRequest(http.MethodPost, "/aws/emr/mrScaler")
	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	gs, err := scalersFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(CreateScalerOutput)
	if len(gs) > 0 {
		output.Scaler = gs[0]
	}

	return output, nil
}

func (s *ServiceOp) Read(ctx context.Context, input *ReadScalerInput) (*ReadScalerOutput, error) {
	path, err := uritemplates.Expand("/aws/emr/mrScaler/{mrScalerId}", uritemplates.Values{
		"mrScalerId": spotinst.StringValue(input.ScalerID),
	})
	if err != nil {
		return nil, err
	}

	r := client.NewRequest(http.MethodGet, path)
	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	gs, err := scalersFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(ReadScalerOutput)
	if len(gs) > 0 {
		output.Scaler = gs[0]
	}

	return output, nil
}

func (s *ServiceOp) Update(ctx context.Context, input *UpdateScalerInput) (*UpdateScalerOutput, error) {
	path, err := uritemplates.Expand("/aws/emr/mrScaler/{mrScalerId}", uritemplates.Values{
		"mrScalerId": spotinst.StringValue(input.Scaler.ID),
	})
	if err != nil {
		return nil, err
	}

	// We do not need the ID anymore so let's drop it.
	input.Scaler.ID = nil

	r := client.NewRequest(http.MethodPut, path)
	r.Obj = input

	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	gs, err := scalersFromHttpResponse(resp)
	if err != nil {
		return nil, err
	}

	output := new(UpdateScalerOutput)
	if len(gs) > 0 {
		output.Scaler = gs[0]
	}

	return output, nil
}

func (s *ServiceOp) Delete(ctx context.Context, input *DeleteScalerInput) (*DeleteScalerOutput, error) {
	path, err := uritemplates.Expand("/aws/emr/mrScaler/{mrScalerId}", uritemplates.Values{
		"mrScalerId": spotinst.StringValue(input.ScalerID),
	})
	if err != nil {
		return nil, err
	}

	r := client.NewRequest(http.MethodDelete, path)
	resp, err := client.RequireOK(s.Client.Do(ctx, r))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return &DeleteScalerOutput{}, nil
}
