package paperspace

import "errors"

type MachineType struct {
	Label string
	CPU   int64
	GPU   int64
	RAM   int64
}

var machineTypes = map[string]*MachineType{
	"Free-CPU":   {Label: "Free-CPU", CPU: 2, GPU: 0, RAM: 2147483648},
	"Free-GPU":   {Label: "Free-GPU", CPU: 8, GPU: 1, RAM: 32212254720},
	"Free-P5000": {Label: "Free-P5000", CPU: 8, GPU: 1, RAM: 32212254720},
	"P4000":      {Label: "P4000", CPU: 8, GPU: 1, RAM: 32212254720},
	"P5000":      {Label: "P5000", CPU: 8, GPU: 1, RAM: 32212254720},
	"P6000":      {Label: "P6000", CPU: 8, GPU: 1, RAM: 32212254720},
	"C1":         {Label: "C1", CPU: 1, GPU: 0, RAM: 536870912},
	"C2":         {Label: "C2", CPU: 1, GPU: 0, RAM: 1073741824},
	"C3":         {Label: "C3", CPU: 2, GPU: 0, RAM: 2147483648},
	"C4":         {Label: "C4", CPU: 2, GPU: 0, RAM: 4294967296},
	"C5":         {Label: "C5", CPU: 4, GPU: 0, RAM: 8589934592},
	"C6":         {Label: "C6", CPU: 8, GPU: 0, RAM: 17179869184},
	"C7":         {Label: "C7", CPU: 12, GPU: 0, RAM: 32212254720},
	"C8":         {Label: "C8", CPU: 16, GPU: 0, RAM: 64424509440},
	"C9":         {Label: "C9", CPU: 24, GPU: 0, RAM: 128849018880},
	"C10":        {Label: "C10", CPU: 32, GPU: 0, RAM: 261993005056},
	"V100":       {Label: "V100", CPU: 8, GPU: 1, RAM: 32212254720},
}

func machineTypeForLabel(machineType string) (*MachineType, error) {
	t, ok := machineTypes[machineType]
	if !ok {
		return nil, errors.New("machine type not found")
	}
	return t, nil
}
