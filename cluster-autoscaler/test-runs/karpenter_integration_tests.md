# Karpenter Integration Tests Execution Log

## Run on 2026-07-10 16:58:39

### Command
`go test -v ./core/scaleup/orchestrator/ -run=TestKarpenter && go test -v ./simulator/karpenter/...`

### Outcome
SUCCESS

### Tests Run and Results
- In `core/scaleup/orchestrator/`:
  - `TestKarpenterSimulatorLabelRequirements` (PASS)
  - `TestKarpenterSimulatorSmallestInstanceTypeResolution` (PASS)
  - `TestKarpenterSimulatorMultiZoneReliabilityBalancing` (PASS)
  - `TestKarpenterSimulatorDaemonSets` (PASS)
  - `TestKarpenterSimulatorClusteringWithDifferentLabels` (PASS)
- In `simulator/karpenter/`:
  - `TestDirectClient_Get_CSINode` (PASS)
  - `TestKarpenterReschedulingSimulator_TrySchedulePods` (PASS)

### Key Test Details

#### `TestDirectClient_Get_CSINode` (New)
Verifies that the `DirectClient` (mock client used by Karpenter) correctly resolves `CSINode` requests:
- **Existing CSINode**: If a `CSINode` exists in the snapshot, it is returned as-is (preserving custom topology keys like `topology.gke.io/zone`).
- **Non-existing / Simulated CSINode**: If a `CSINode` is not found (e.g. for simulated nodes, or if CSI is disabled), the client falls back to generating a mock `CSINode`.
  - The mock `CSINode` contains drivers translated from the `StorageClass`es present in the snapshot (using `csi-translation-lib` to translate in-tree provisioners like `kubernetes.io/gce-pd` to `pd.csi.storage.gke.io`).
  - The mock `CSINode` is populated with standard topology keys (`topology.kubernetes.io/zone` and `kubernetes.io/hostname`), ensuring Karpenter can resolve zonal constraints for PVs.

---

## Run on 2026-07-10 16:51:06

### Command
`go test -v ./core/scaleup/orchestrator/...`

### Outcome
SUCCESS

### Tests Run and Results
- `TestKarpenterSimulatorLabelRequirements` (PASS)
- `TestKarpenterSimulatorSmallestInstanceTypeResolution` (PASS)
- `TestKarpenterSimulatorMultiZoneReliabilityBalancing` (PASS)
- `TestKarpenterSimulatorDaemonSets` (PASS)
- `TestKarpenterSimulatorClusteringWithDifferentLabels` (PASS)
- All other orchestrator tests (PASS)

---

## Run on 2026-07-10 16:13:41

### Command
`go test -v ./core/scaleup/orchestrator/...`

### Outcome
SUCCESS

### Tests Run and Results
- `TestKarpenterSimulatorLabelRequirements` (PASS)
- `TestKarpenterSimulatorSmallestInstanceTypeResolution` (PASS)
- `TestKarpenterSimulatorMultiZoneReliabilityBalancing` (PASS)
- `TestKarpenterSimulatorDaemonSets` (PASS)
- `TestKarpenterSimulatorClusteringWithDifferentLabels` (PASS)
- All other orchestrator tests (PASS)

### Key Test Details (Updated to GCP Naming)

#### `TestKarpenterSimulatorDaemonSets`
Verifies that when two node groups are split into different virtual `InstanceType`s (due to different DaemonSet requirements), they are correctly routed.
Specifically:
- `ng-ds` has a DaemonSet `ds1` (200 CPU).
- `ng-nods` has no DaemonSets.
- Both use physical IT `e2-standard-2` (labeled as such, 1000 CPU in test template).
- `pod1` requires 900 CPU.
- `ng-ds` virtual IT (`e2-standard-2-1`) requirements homogeneous labels includes `ng: ng-ds`.
- `ng-nods` virtual IT (`e2-standard-2`) requirements homogeneous labels includes `ng: ng-nods`.
- `ds1` requires `ng: ng-ds`.
- Karpenter correctly schedules `ds1` only on `e2-standard-2-1`.
- Karpenter correctly schedules `pod1` on `e2-standard-2` (`ng-nods`) because it has 0 DS overhead and 1000 CPU available, whereas `e2-standard-2-1` has 200 CPU overhead and only 800 CPU available (so `pod1` doesn't fit).
- The test successfully asserts that `ng-nods` was chosen for scale-up.

#### `TestKarpenterSimulatorClusteringWithDifferentLabels`
Verifies that when two node groups share the same physical IT and DaemonSets, but have different custom labels, they are clustered into the SAME virtual `InstanceType`.
Specifically:
- `ng1` has label `custom-label: val1`.
- `ng2` has label `custom-label: val2`.
- Both use `e2-standard-2` and have no DaemonSets.
- They are clustered into one `InstanceType` `e2-standard-2`.
- `custom-label` is NOT in `e2-standard-2` IT requirements (since it differs, it is not homogeneous).
- `custom-label` is in `Offering` requirements: `ng1` offering has `custom-label: val1`, `ng2` offering has `custom-label: val2`.
- `pod1` requires `custom-label: val2`.
- Karpenter correctly schedules `pod1` on `e2-standard-2` with `ng2` offering.
- The simulator correctly resolves the claim to `ng2` using `ngMatchesRequirements`.
- The test successfully asserts that `ng2` was chosen for scale-up.
