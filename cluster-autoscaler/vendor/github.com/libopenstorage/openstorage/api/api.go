package api

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes"

	"github.com/mohae/deepcopy"
)

// Strings for VolumeSpec
const (
	Name                     = "name"
	SpecNodes                = "nodes"
	SpecParent               = "parent"
	SpecEphemeral            = "ephemeral"
	SpecShared               = "shared"
	SpecJournal              = "journal"
	SpecSharedv4             = "sharedv4"
	SpecCascaded             = "cascaded"
	SpecSticky               = "sticky"
	SpecSecure               = "secure"
	SpecCompressed           = "compressed"
	SpecSize                 = "size"
	SpecScale                = "scale"
	SpecFilesystem           = "fs"
	SpecBlockSize            = "block_size"
	SpecHaLevel              = "repl"
	SpecPriority             = "io_priority"
	SpecSnapshotInterval     = "snap_interval"
	SpecSnapshotSchedule     = "snap_schedule"
	SpecAggregationLevel     = "aggregation_level"
	SpecDedupe               = "dedupe"
	SpecPassphrase           = "secret_key"
	SpecAutoAggregationValue = "auto"
	SpecGroup                = "group"
	SpecGroupEnforce         = "fg"
	SpecZones                = "zones"
	SpecRacks                = "racks"
	SpecRack                 = "rack"
	SpecRegions              = "regions"
	SpecLabels               = "labels"
	SpecPriorityAlias        = "priority_io"
	SpecIoProfile            = "io_profile"
)

// OptionKey specifies a set of recognized query params.
const (
	// OptName query parameter used to lookup volume by name.
	OptName = "Name"
	// OptVolumeID query parameter used to lookup volume by ID.
	OptVolumeID = "VolumeID"
	// OptSnapID query parameter used to lookup snapshot by ID.
	OptSnapID = "SnapID"
	// OptLabel query parameter used to lookup volume by set of labels.
	OptLabel = "Label"
	// OptConfigLabel query parameter used to lookup volume by set of labels.
	OptConfigLabel = "ConfigLabel"
	// OptCumulative query parameter used to request cumulative stats.
	OptCumulative = "Cumulative"
	// OptTimeout query parameter used to indicate timeout seconds
	OptTimeoutSec = "TimeoutSec"
	// OptQuiesceID query parameter use for quiesce
	OptQuiesceID = "QuiesceID"
	// OptCredUUID is the UUID of the credential
	OptCredUUID = "CredUUID"
	// OptCredType  indicates type of credential
	OptCredType = "CredType"
	// OptCredEncrKey is the key used to encrypt data
	OptCredEncrKey = "CredEncrypt"
	// OptCredRegion indicates the region for s3
	OptCredRegion = "CredRegion"
	// OptCredDisableSSL indicated if SSL should be disabled
	OptCredDisableSSL = "CredDisableSSL"
	// OptCredEndpoint indicate the cloud endpoint
	OptCredEndpoint = "CredEndpoint"
	// OptCredAccKey for s3
	OptCredAccessKey = "CredAccessKey"
	// OptCredSecretKey for s3
	OptCredSecretKey = "CredSecretKey"
	// OptCredGoogleProjectID projectID for google cloud
	OptCredGoogleProjectID = "CredProjectID"
	// OptCredGoogleJsonKey for google cloud
	OptCredGoogleJsonKey = "CredJsonKey"
	// OptCredAzureAccountName is the account name for
	// azure as the cloud provider
	OptCredAzureAccountName = "CredAccountName"
	// OptOptCredAzureAccountKey is the accountkey for
	// azure as the cloud provider
	OptCredAzureAccountKey = "CredAccountKey"
	// OptCloudBackupID is the backID in the cloud
	OptCloudBackupID = "CloudBackID"
	// OptSrcVolID is the source volume ID of the backup
	OptSrcVolID = "SrcVolID"
	// OptBkupOpState is the desired operational state
	// (stop/pause/resume) of backup/restore
	OptBkupOpState = "OpState"
	// OptBackupSchedUUID is the UUID of the backup-schedule
	OptBackupSchedUUID = "BkupSchedUUID"
)

// Api clientserver Constants
const (
	OsdVolumePath        = "osd-volumes"
	OsdSnapshotPath      = "osd-snapshot"
	OsdCredsPath         = "osd-creds"
	OsdBackupPath        = "osd-backup"
	OsdMigratePath       = "osd-migrate"
	OsdMigrateStartPath  = OsdMigratePath + "/start"
	OsdMigrateCancelPath = OsdMigratePath + "/cancel"
	OsdMigrateStatusPath = OsdMigratePath + "/status"
	TimeLayout           = "Jan 2 15:04:05 UTC 2006"
)

const (
	// AutoAggregation value indicates driver to select aggregation level.
	AutoAggregation = math.MaxUint32
)

// Node describes the state of a node.
// It includes the current physical state (CPU, memory, storage, network usage) as
// well as the containers running on the system.
//
// swagger:model
type Node struct {
	// Id of the node.
	Id string
	// Cpu usage of the node.
	Cpu float64 // percentage.
	// Total Memory of the node
	MemTotal uint64
	// Used Memory of the node
	MemUsed uint64
	// Free Memory of the node
	MemFree uint64
	// Average load (percentage)
	Avgload int
	// Node Status see (Status object)
	Status Status
	// GenNumber of the node
	GenNumber uint64
	// List of disks on this node.
	Disks map[string]StorageResource
	// List of storage pools this node supports
	Pools []StoragePool
	// Management IP
	MgmtIp string
	// Data IP
	DataIp string
	// Timestamp
	Timestamp time.Time
	// Start time of this node
	StartTime time.Time
	// Hostname of this node
	Hostname string
	// Node data for this node (EX: Public IP, Provider, City..)
	NodeData map[string]interface{}
	// User defined labels for node. Key Value pairs
	NodeLabels map[string]string
}

// FluentDConfig describes ip and port of a fluentdhost.
// DEPRECATED
//
// swagger:model
type FluentDConfig struct {
	IP   string `json:"ip"`
	Port string `json:"port"`
}

// TunnelConfig describes key, cert and endpoint of a reverse proxy tunnel
// DEPRECATED
//
// swagger:model
type TunnelConfig struct {
	Key      string `json:"key"`
	Cert     string `json:"cert"`
	Endpoint string `json:"tunnel_endpoint"`
}

// Cluster represents the state of the cluster.
//
// swagger:model
type Cluster struct {
	Status Status

	// Id of the cluster.
	//
	// required: true
	Id string

	// Id of the node on which this cluster object is initialized
	NodeId string

	// array of all the nodes in the cluster.
	Nodes []Node

	// Logging url for the cluster.
	LoggingURL string

	// Management url for the cluster
	ManagementURL string

	// FluentD Host for the cluster
	FluentDConfig FluentDConfig

	// TunnelConfig for the cluster [key, cert, endpoint]
	TunnelConfig TunnelConfig
}

// CredCreateRequest is the input for CredCreate command
type CredCreateRequest struct {
	// InputParams is map describing cloud provide
	InputParams map[string]string
}

// CredCreateResponse is returned for CredCreate command
type CredCreateResponse struct {
	// UUID of the credential that was just created
	UUID string
}

// StatPoint represents the basic structure of a single Stat reported
// TODO: This is the first step to introduce stats in openstorage.
//       Follow up task is to introduce an API for logging stats
type StatPoint struct {
	// Name of the Stat
	Name string
	// Tags for the Stat
	Tags map[string]string
	// Fields and values of the stat
	Fields map[string]interface{}
	// Timestamp in Unix format
	Timestamp int64
}

type CloudBackupCreateRequest struct {
	// VolumeID of the volume for which cloudbackup is requested
	VolumeID string
	// CredentialUUID is cloud credential to be used for backup
	CredentialUUID string
	// Full indicates if full backup is desired even though incremental is possible
	Full bool
}

type CloudBackupRestoreRequest struct {
	// ID is the backup ID being restored
	ID string
	// RestoreVolumeName is optional volume Name of the new volume to be created
	// in the cluster for restoring the cloudbackup
	RestoreVolumeName string
	// CredentialUUID is the credential to be used for restore operation
	CredentialUUID string
	// NodeID is the optional NodeID for provisioning restore
	// volume (ResoreVolumeName should not be specified)
	NodeID string
}

type CloudBackupRestoreResponse struct {
	// RestoreVolumeID is the volumeID to which the backup is being restored
	RestoreVolumeID string
}

type CloudBackupGenericRequest struct {
	// SrcVolumeID is optional Source VolumeID for the request
	SrcVolumeID string
	// ClusterID is the optional clusterID for the request
	ClusterID string
	// CredentialUUID is the credential for cloud to be used for the request
	CredentialUUID string
	// All if set to true, backups for all clusters in the cloud are processed
	All bool
}

type CloudBackupInfo struct {
	// ID is the ID of the cloud backup
	ID string
	// SrcVolumeID is Source volumeID of the backup
	SrcVolumeID string
	// SrcvolumeName is name of the sourceVolume of the backup
	SrcVolumeName string
	// Timestamp is the timestamp at which the source volume
	// was backed up to cloud
	Timestamp time.Time
	// Metadata associated with the backup
	Metadata map[string]string
	// Status indicates the status of the backup
	Status string
}

type CloudBackupEnumerateRequest struct {
	CloudBackupGenericRequest
}

type CloudBackupEnumerateResponse struct {
	// Backups is list of backups in cloud for given volume/cluster/s
	Backups []CloudBackupInfo
}

type CloudBackupDeleteRequest struct {
	// ID is the ID of the cloud backup
	ID string
	// CredentialUUID is the credential for cloud to be used for the request
	CredentialUUID string
	// Force Delete cloudbackup even if there are dependencies
	Force bool
}

type CloudBackupDeleteAllRequest struct {
	CloudBackupGenericRequest
}

type CloudBackupStatusRequest struct {
	// SrcVolumeID optional volumeID to list status of backup/restore
	SrcVolumeID string
	// Local indicates if only those backups/restores that are
	// active on current node must be returned
	Local bool
}

type CloudBackupOpType string

const (
	CloudBackupOp  = CloudBackupOpType("Backup")
	CloudRestoreOp = CloudBackupOpType("Restore")
)

type CloudBackupStatusType string

const (
	CloudBackupStatusNotStarted = CloudBackupStatusType("NotStarted")
	CloudBackupStatusDone       = CloudBackupStatusType("Done")
	CloudBackupStatusAborted    = CloudBackupStatusType("Aborted")
	CloudBackupStatusPaused     = CloudBackupStatusType("Paused")
	CloudBackupStatusStopped    = CloudBackupStatusType("Stopped")
	CloudBackupStatusActive     = CloudBackupStatusType("Active")
	CloudBackupStatusFailed     = CloudBackupStatusType("Failed")
)

const (
	CloudBackupRequestedStatePause  = "pause"
	CloudBackupRequestedStateResume = "resume"
	CloudBackupRequestedStateStop   = "stop"
)

type CloudBackupStatus struct {
	// ID is the ID for the operation
	ID string
	// OpType indicates if this is a backup or restore
	OpType CloudBackupOpType
	// State indicates if the op is currently active/done/failed
	Status CloudBackupStatusType
	// BytesDone indicates total Bytes uploaded/downloaded
	BytesDone uint64
	// StartTime indicates Op's start time
	StartTime time.Time
	// CompletedTime indicates Op's completed time
	CompletedTime time.Time
	// NodeID is the ID of the node where this Op is active
	NodeID string
}

type CloudBackupStatusResponse struct {
	// statuses is list of currently active/failed/done backup/restores
	Statuses map[string]CloudBackupStatus
}

type CloudBackupCatalogRequest struct {
	// ID is Backup ID in the cloud
	ID string
	// CredentialUUID is the credential for cloud
	CredentialUUID string
}

type CloudBackupCatalogResponse struct {
	// Contents is listing of backup contents
	Contents []string
}

type CloudBackupHistoryRequest struct {
	// SrcVolumeID is volumeID for which history of backup/restore
	// is being requested
	SrcVolumeID string
}

type CloudBackupHistoryItem struct {
	// SrcVolumeID is volume ID which was backedup
	SrcVolumeID string
	// TimeStamp is the time at which either backup completed/failed
	Timestamp time.Time
	// Status indicates whether backup was completed/failed
	Status string
}

type CloudBackupHistoryResponse struct {
	// HistoryList is list of past backup/restores in the cluster
	HistoryList []CloudBackupHistoryItem
}

type CloudBackupStateChangeRequest struct {
	// SrcVolumeID is volume ID on which backup/restore
	// state change is being requested
	SrcVolumeID string
	// RequestedState is desired state of the op
	// can be pause/resume/stop
	RequestedState string
}

type CloudBackupScheduleInfo struct {
	// SrcVolumeID is the schedule's source volume
	SrcVolumeID string
	// CredentialUUID is the cloud credential used with this schedule
	CredentialUUID string
	// Schedule is the frequence of backup
	Schedule string
	// MaxBackups are the maximum number of backups retained
	// in cloud.Older backups are deleted
	MaxBackups uint
}

type CloudBackupSchedCreateRequest struct {
	CloudBackupScheduleInfo
}

type CloudBackupSchedCreateResponse struct {
	// UUID is the UUID of the newly created schedule
	UUID string
}

type CloudBackupSchedDeleteRequest struct {
	// UUID is UUID of the schedule to be deleted
	UUID string
}

type CloudBackupSchedEnumerateResponse struct {
	// Schedule is map of schedule uuid to scheduleInfo
	Schedules map[string]CloudBackupScheduleInfo
}

// DriverTypeSimpleValueOf returns the string format of DriverType
func DriverTypeSimpleValueOf(s string) (DriverType, error) {
	obj, err := simpleValueOf("driver_type", DriverType_value, s)
	return DriverType(obj), err
}

// SimpleString returns the string format of DriverType
func (x DriverType) SimpleString() string {
	return simpleString("driver_type", DriverType_name, int32(x))
}

// FSTypeSimpleValueOf returns the string format of FSType
func FSTypeSimpleValueOf(s string) (FSType, error) {
	obj, err := simpleValueOf("fs_type", FSType_value, s)
	return FSType(obj), err
}

// SimpleString returns the string format of DriverType
func (x FSType) SimpleString() string {
	return simpleString("fs_type", FSType_name, int32(x))
}

// CosTypeSimpleValueOf returns the string format of CosType
func CosTypeSimpleValueOf(s string) (CosType, error) {
	obj, exists := CosType_value[strings.ToUpper(s)]
	if !exists {
		return -1, fmt.Errorf("Invalid cos value: %s", s)
	}
	return CosType(obj), nil
}

// SimpleString returns the string format of CosType
func (x CosType) SimpleString() string {
	return simpleString("cos_type", CosType_name, int32(x))
}

// GraphDriverChangeTypeSimpleValueOf returns the string format of GraphDriverChangeType
func GraphDriverChangeTypeSimpleValueOf(s string) (GraphDriverChangeType, error) {
	obj, err := simpleValueOf("graph_driver_change_type", GraphDriverChangeType_value, s)
	return GraphDriverChangeType(obj), err
}

// SimpleString returns the string format of GraphDriverChangeType
func (x GraphDriverChangeType) SimpleString() string {
	return simpleString("graph_driver_change_type", GraphDriverChangeType_name, int32(x))
}

// VolumeActionParamSimpleValueOf returns the string format of VolumeAction
func VolumeActionParamSimpleValueOf(s string) (VolumeActionParam, error) {
	obj, err := simpleValueOf("volume_action_param", VolumeActionParam_value, s)
	return VolumeActionParam(obj), err
}

// SimpleString returns the string format of VolumeAction
func (x VolumeActionParam) SimpleString() string {
	return simpleString("volume_action_param", VolumeActionParam_name, int32(x))
}

// VolumeStateSimpleValueOf returns the string format of VolumeState
func VolumeStateSimpleValueOf(s string) (VolumeState, error) {
	obj, err := simpleValueOf("volume_state", VolumeState_value, s)
	return VolumeState(obj), err
}

// SimpleString returns the string format of VolumeState
func (x VolumeState) SimpleString() string {
	return simpleString("volume_state", VolumeState_name, int32(x))
}

// VolumeStatusSimpleValueOf returns the string format of VolumeStatus
func VolumeStatusSimpleValueOf(s string) (VolumeStatus, error) {
	obj, err := simpleValueOf("volume_status", VolumeStatus_value, s)
	return VolumeStatus(obj), err
}

// SimpleString returns the string format of VolumeStatus
func (x VolumeStatus) SimpleString() string {
	return simpleString("volume_status", VolumeStatus_name, int32(x))
}

// IoProfileSimpleValueOf returns the string format of IoProfile
func IoProfileSimpleValueOf(s string) (IoProfile, error) {
	obj, err := simpleValueOf("io_profile", IoProfile_value, s)
	return IoProfile(obj), err
}

// SimpleString returns the string format of IoProfile
func (x IoProfile) SimpleString() string {
	return simpleString("io_profile", IoProfile_name, int32(x))
}

func simpleValueOf(typeString string, valueMap map[string]int32, s string) (int32, error) {
	obj, ok := valueMap[strings.ToUpper(fmt.Sprintf("%s_%s", typeString, s))]
	if !ok {
		return 0, fmt.Errorf("no openstorage.%s for %s", strings.ToUpper(typeString), s)
	}
	return obj, nil
}

func simpleString(typeString string, nameMap map[int32]string, v int32) string {
	s, ok := nameMap[v]
	if !ok {
		return strconv.Itoa(int(v))
	}
	return strings.TrimPrefix(strings.ToLower(s), fmt.Sprintf("%s_", strings.ToLower(typeString)))
}

func toSec(ms uint64) uint64 {
	return ms / 1000
}

// WriteThroughput returns the write throughput
func (v *Stats) WriteThroughput() uint64 {
	intv := toSec(v.IntervalMs)
	if intv == 0 {
		return 0
	}
	return (v.WriteBytes) / intv
}

// ReadThroughput returns the read throughput
func (v *Stats) ReadThroughput() uint64 {
	intv := toSec(v.IntervalMs)
	if intv == 0 {
		return 0
	}
	return (v.ReadBytes) / intv
}

// Latency returns latency
func (v *Stats) Latency() uint64 {
	ops := v.Writes + v.Reads
	if ops == 0 {
		return 0
	}
	return (uint64)((v.IoMs * 1000) / ops)
}

// Read latency returns avg. time required for read operation to complete
func (v *Stats) ReadLatency() uint64 {
	if v.Reads == 0 {
		return 0
	}
	return (uint64)((v.ReadMs * 1000) / v.Reads)
}

// Write latency returns avg. time required for write operation to complete
func (v *Stats) WriteLatency() uint64 {
	if v.Writes == 0 {
		return 0
	}
	return (uint64)((v.WriteMs * 1000) / v.Writes)
}

// Iops returns iops
func (v *Stats) Iops() uint64 {
	intv := toSec(v.IntervalMs)
	if intv == 0 {
		return 0
	}
	return (v.Writes + v.Reads) / intv
}

// Scaled returns true if the volume is scaled.
func (v *Volume) Scaled() bool {
	return v.Spec.Scale > 1
}

// Contains returns true if mid is a member of volume's replication set.
func (m *Volume) Contains(mid string) bool {
	rsets := m.GetReplicaSets()
	for _, rset := range rsets {
		for _, node := range rset.Nodes {
			if node == mid {
				return true
			}
		}
	}
	return false
}

// Copy makes a deep copy of VolumeSpec
func (s *VolumeSpec) Copy() *VolumeSpec {
	spec := *s
	if s.VolumeLabels != nil {
		spec.VolumeLabels = make(map[string]string)
		for k, v := range s.VolumeLabels {
			spec.VolumeLabels[k] = v
		}
	}
	if s.ReplicaSet != nil {
		spec.ReplicaSet = &ReplicaSet{Nodes: make([]string, len(s.ReplicaSet.Nodes))}
		copy(spec.ReplicaSet.Nodes, s.ReplicaSet.Nodes)
	}
	return &spec
}

// Copy makes a deep copy of Node
func (s *Node) Copy() *Node {
	localCopy := deepcopy.Copy(*s)
	nodeCopy := localCopy.(Node)
	return &nodeCopy
}

func (v Volume) IsClone() bool {
	return v.Source != nil && len(v.Source.Parent) != 0 && !v.Readonly
}

func (v Volume) IsSnapshot() bool {
	return v.Source != nil && len(v.Source.Parent) != 0 && v.Readonly
}

func (v Volume) DisplayId() string {
	if v.Locator != nil {
		return fmt.Sprintf("%s (%s)", v.Locator.Name, v.Id)
	} else {
		return v.Id
	}
}

// ToStorageNode converts a Node structure to an exported gRPC StorageNode struct
func (s *Node) ToStorageNode() *StorageNode {
	node := &StorageNode{
		Id:       s.Id,
		Cpu:      s.Cpu,
		MemTotal: s.MemTotal,
		MemUsed:  s.MemUsed,
		MemFree:  s.MemFree,
		AvgLoad:  int64(s.Avgload),
		Status:   s.Status,
		MgmtIp:   s.MgmtIp,
		DataIp:   s.DataIp,
		Hostname: s.Hostname,
	}

	node.Disks = make(map[string]*StorageResource)
	for k, v := range s.Disks {
		node.Disks[k] = &v
	}

	node.NodeLabels = make(map[string]string)
	for k, v := range s.NodeLabels {
		node.NodeLabels[k] = v
	}

	node.Pools = make([]*StoragePool, len(s.Pools))
	for i, v := range s.Pools {
		node.Pools[i] = &v
	}

	return node
}

// ToStorageCluster converts a Cluster structure to an exported gRPC StorageCluster struct
func (c *Cluster) ToStorageCluster() *StorageCluster {
	cluster := &StorageCluster{
		Status: c.Status,
		Id:     c.Id,
	}

	return cluster
}

func CloudBackupStatusTypeToSdkCloudBackupStatusType(
	t CloudBackupStatusType,
) SdkCloudBackupStatusType {
	switch t {
	case CloudBackupStatusNotStarted:
		return SdkCloudBackupStatusType_SdkCloudBackupStatusTypeNotStarted
	case CloudBackupStatusDone:
		return SdkCloudBackupStatusType_SdkCloudBackupStatusTypeDone
	case CloudBackupStatusAborted:
		return SdkCloudBackupStatusType_SdkCloudBackupStatusTypeAborted
	case CloudBackupStatusPaused:
		return SdkCloudBackupStatusType_SdkCloudBackupStatusTypePaused
	case CloudBackupStatusStopped:
		return SdkCloudBackupStatusType_SdkCloudBackupStatusTypeStopped
	case CloudBackupStatusActive:
		return SdkCloudBackupStatusType_SdkCloudBackupStatusTypeActive
	case CloudBackupStatusFailed:
		return SdkCloudBackupStatusType_SdkCloudBackupStatusTypeFailed
	default:
		return SdkCloudBackupStatusType_SdkCloudBackupStatusTypeUnknown
	}
}

func StringToSdkCloudBackupStatusType(s string) SdkCloudBackupStatusType {
	return CloudBackupStatusTypeToSdkCloudBackupStatusType(CloudBackupStatusType(s))
}

func (b *CloudBackupInfo) ToSdkCloudBackupInfo() *SdkCloudBackupInfo {
	info := &SdkCloudBackupInfo{
		Id:            b.ID,
		SrcVolumeId:   b.SrcVolumeID,
		SrcVolumeName: b.SrcVolumeName,
		Metadata:      b.Metadata,
	}

	info.Timestamp, _ = ptypes.TimestampProto(b.Timestamp)
	info.Status = StringToSdkCloudBackupStatusType(b.Status)

	return info
}

func (r *CloudBackupEnumerateResponse) ToSdkCloudBackupEnumerateResponse() *SdkCloudBackupEnumerateResponse {
	resp := &SdkCloudBackupEnumerateResponse{
		Backups: make([]*SdkCloudBackupInfo, len(r.Backups)),
	}

	for i, v := range r.Backups {
		resp.Backups[i] = v.ToSdkCloudBackupInfo()
	}

	return resp
}

func CloudBackupOpTypeToSdkCloudBackupOpType(t CloudBackupOpType) SdkCloudBackupOpType {
	switch t {
	case CloudBackupOp:
		return SdkCloudBackupOpType_SdkCloudBackupOpTypeBackupOp
	case CloudRestoreOp:
		return SdkCloudBackupOpType_SdkCloudBackupOpTypeRestoreOp
	default:
		return SdkCloudBackupOpType_SdkCloudBackupOpTypeUnknown
	}
}

func StringToSdkCloudBackupOpType(s string) SdkCloudBackupOpType {
	return CloudBackupOpTypeToSdkCloudBackupOpType(CloudBackupOpType(s))
}

func (s CloudBackupStatus) ToSdkCloudBackupStatus() *SdkCloudBackupStatus {
	status := &SdkCloudBackupStatus{
		BackupId:  s.ID,
		Optype:    CloudBackupOpTypeToSdkCloudBackupOpType(s.OpType),
		Status:    CloudBackupStatusTypeToSdkCloudBackupStatusType(s.Status),
		BytesDone: s.BytesDone,
		NodeId:    s.NodeID,
	}

	status.StartTime, _ = ptypes.TimestampProto(s.StartTime)
	status.CompletedTime, _ = ptypes.TimestampProto(s.CompletedTime)

	return status
}

func (r *CloudBackupStatusResponse) ToSdkCloudBackupStatusResponse() *SdkCloudBackupStatusResponse {
	resp := &SdkCloudBackupStatusResponse{
		Statuses: make(map[string]*SdkCloudBackupStatus),
	}

	for k, v := range r.Statuses {
		resp.Statuses[k] = v.ToSdkCloudBackupStatus()
	}

	return resp
}

func (h CloudBackupHistoryItem) ToSdkCloudBackupHistoryItem() *SdkCloudBackupHistoryItem {
	item := &SdkCloudBackupHistoryItem{
		SrcVolumeId: h.SrcVolumeID,
		Status:      StringToSdkCloudBackupStatusType(h.Status),
	}

	item.Timestamp, _ = ptypes.TimestampProto(h.Timestamp)
	return item
}

func (r *CloudBackupHistoryResponse) ToSdkCloudBackupHistoryResponse() *SdkCloudBackupHistoryResponse {
	resp := &SdkCloudBackupHistoryResponse{
		HistoryList: make([]*SdkCloudBackupHistoryItem, len(r.HistoryList)),
	}

	for i, v := range r.HistoryList {
		resp.HistoryList[i] = v.ToSdkCloudBackupHistoryItem()
	}

	return resp
}
