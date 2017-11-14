package flex

import (
	"encoding/json"
	"fmt"

	"k8s.io/kubernetes/pkg/volume/flexvolume"
)

type VolumeResult flexvolume.DriverStatus

const (
	StatusNotSupported = flexvolume.StatusNotSupported
	StatusSuccess      = flexvolume.StatusSuccess
	StatusFailure      = "Failure"
)

func (v VolumeResult) ToJson() string {
	ret, _ := json.Marshal(&v)
	return string(ret)
}

func (v VolumeResult) WithDevicePath(devicePath string) VolumeResult {
	v.DevicePath = devicePath
	return v
}

func (v VolumeResult) WithVolumeName(volumeName string) VolumeResult {
	v.VolumeName = volumeName
	return v
}

func (v VolumeResult) WithV(devicePath string) VolumeResult {
	v.DevicePath = devicePath
	return v
}

func (v VolumeResult) WithAttached(attached bool) VolumeResult {
	v.Attached = attached
	return v
}

func (v VolumeResult) Error() string {
	return v.Message
}

// NewVolumeError creates failure error with given message
func NewVolumeError(msg string, args ...interface{}) VolumeResult {
	return VolumeResult{Message: fmt.Sprintf(msg, args...), Status: StatusFailure}
}

func NewVolumeNotSupported(msg string) VolumeResult {
	return VolumeResult{Message: msg, Status: StatusNotSupported}
}

func NewVolumeSuccess() VolumeResult {
	return VolumeResult{Status: StatusSuccess}
}

type VolumeOptions map[string]interface{}

type VolumePlugin interface {
	Init() VolumeResult
	Attach(options VolumeOptions, node string) VolumeResult
	Detach(pvOrVolumeName string, node string) VolumeResult
	MountDevice(dir string, _, options VolumeOptions) VolumeResult
	UnmountDevice(dir string) VolumeResult
	WaitForAttach(_, options VolumeOptions) VolumeResult
	GetVolumeName(options VolumeOptions) VolumeResult
	IsAttached(options VolumeOptions, node string) VolumeResult
}
