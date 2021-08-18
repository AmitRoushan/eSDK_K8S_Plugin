package driver

import(
	"utils/k8sutils"
)

type Driver struct {
	name            string
	version         string
	useMultiPath    bool
	isNeedMultiPath bool
	k8sUtils        k8sutils.Interface
	nodeName 		string
}

func NewDriver(name, version string, useMultiPath, isNeedMultiPath bool,
				k8sUtils k8sutils.Interface, nodeName string) *Driver {
	return &Driver{
		name:            name,
		version:         version,
		useMultiPath:    useMultiPath,
		isNeedMultiPath: isNeedMultiPath,
		k8sUtils:		k8sUtils,
		nodeName: 		nodeName,
	}
}
