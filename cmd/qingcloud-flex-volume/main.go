package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/yunify/qingcloud-volume-provisioner/pkg/volume/flex"
	"github.com/yunify/qingcloud-volume-provisioner/pkg/volume/qingcloud"
	qclogger "github.com/yunify/qingcloud-sdk-go/logger"
	"os"
	"path"
	"path/filepath"
	"github.com/golang/glog"
)

const (
	DriverDir  = "/usr/libexec/kubernetes/kubelet-plugins/volume/exec/"
	LogDir = "/var/log/qingcloud-flex-volume"
)

// fatalf is a convenient method that outputs error in flex volume plugin style
// and quits
func fatalf(msg string, args ...interface{}) {
	err := flex.VolumeResult{
		Message: fmt.Sprintf(msg, args...),
		Status:  "Failure",
	}
	fmt.Printf(err.ToJson())
	os.Exit(1)
}

// printResult is a convenient method for printing result of volume operation
func printResult(result flex.VolumeResult) {
	fmt.Printf(result.ToJson())
	glog.Infof("Response: %#v", result.ToJson())
	if result.Status == "Success" {
		os.Exit(0)
	}
	os.Exit(1)
}

// ensureVolumeOptions decodes json or die
func ensureVolumeOptions(v string) (vo flex.VolumeOptions) {
	err := json.Unmarshal([]byte(v), &vo)
	if err != nil {
		fatalf("Invalid json options: %s", v)
	}
	return
}

func installDriver(){
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	vendor, driver := path.Split(qingcloud.FlexDriverName)
	vendor = path.Clean(vendor)
	driverTargetDir := path.Join(DriverDir, fmt.Sprintf("%s~%s", vendor, driver))
	driverTargetFile := path.Join(driverTargetDir, driver)
	fmt.Printf("Install driver to %s \n", driverTargetFile)
	err = os.MkdirAll(driverTargetDir, 0644)
	if err != nil {
		panic(err)
	}
	if _, err := os.Stat(driverTargetFile); !os.IsNotExist(err) {
		if err = os.Remove(driverTargetFile); err != nil {
			panic(err)
		}
	}
	err = os.Link(ex, driverTargetFile)
	if err != nil {
		panic(err)
	}
}

func main() {
	install := flag.Bool("install", false, fmt.Sprintf("Install %s to %s", qingcloud.FlexDriverName, DriverDir))

	// Prepare logs
	err := os.MkdirAll(LogDir, 0750)
	if err != nil {
		panic(fmt.Sprintf("mkdir %s err: %s", LogDir, err.Error()))
	}

	flag.Set("logtostderr", "false")
	flag.Set("alsologtostderr", "false")
	flag.Set("log_dir", LogDir)
	flag.Set("stderrthreshold", "FATAL")
	flag.Set("v", "4")
	flag.Parse()
	defer glog.Flush()

	glog.Infof("Call %s driver, args: %#v", qingcloud.FlexDriverName,  flag.Args())

	qcSDKLogFile := filepath.Join(LogDir, "qingcloud_sdk.log")
	f, err := os.OpenFile(qcSDKLogFile, os.O_WRONLY | os.O_CREATE, 0755)
	if err != nil {
		fatalf("Error open log file: %s", qcSDKLogFile)
	}
	defer f.Close()

	qclogger.SetOutput(f)

	if *install {
		flag.VisitAll(func(f *flag.Flag) {
			glog.Infof("Flag: %s=%s", f.Name, f.Value)
		})
		installDriver()
		return
	}

	volumePlugin, err := qingcloud.NewFlexVolumePlugin()

	if err != nil {
		fatalf("Error init FlexVolumePlugin")
	}

	args := flag.Args()
	if len(args) == 0 {
		fatalf("Usage: %s init|attach|detach|mountdevice|unmountdevice|waitforattach|getvolumename|isattached", os.Args[0])
	}

	var ret flex.VolumeResult
	op := args[0]
	args = args[1:]
	switch op {
	case "init":
		ret = volumePlugin.Init()
	case "attach":
		if len(args) < 2 {
			fatalf("attach requires options in json format and a node name")
		}
		ret = volumePlugin.Attach(ensureVolumeOptions(args[0]), args[1])
	case "isattached":
		if len(args) < 2 {
			fatalf("isattached requires options in json format and a node name")
		}
		ret = volumePlugin.Attach(ensureVolumeOptions(args[0]), args[1])
	case "detach":
		if len(args) < 2 {
			fatalf("detach requires a device path and a node name")
		}
		ret = volumePlugin.Detach(args[0], args[1])
	case "mountdevice":
		if len(args) < 3 {
			fatalf("mountdevice requires a mount path, a device path and mount options")
		}
		ret = volumePlugin.MountDevice(args[0], args[1], ensureVolumeOptions(args[2]))
	case "unmountdevice":
		if len(args) < 1 {
			fatalf("unmountdevice requires a mount path")
		}
		ret = volumePlugin.UnmountDevice(args[0])
	case "waitforattach":
		if len(args) < 2 {
			fatalf("waitforattach requires a device path and options in json format")
		}
		ret = volumePlugin.WaitForAttach(args[0], ensureVolumeOptions(args[1]))
	case "getvolumename":
		if len(args) < 1 {
			fatalf("getvolumename requires options in json format")
		}
		ret = volumePlugin.GetVolumeName(ensureVolumeOptions(args[0]))
	default:
		ret = flex.NewVolumeNotSupported(op)
	}

	printResult(ret)
}
