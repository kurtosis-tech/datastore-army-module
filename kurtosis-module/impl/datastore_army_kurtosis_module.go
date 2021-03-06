package impl

import (
	"encoding/json"
	"fmt"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis-core-api-lib/api/golang/lib/services"
	"github.com/kurtosis-tech/stacktrace"
)

const (
	datastoreImage = "kurtosistech/example-datastore-server"

	datastorePortId  = "grpc"
	datastorePortNum = uint16(1323)
)

var datastorePortSpec = services.NewPortSpec(datastorePortNum, services.PortProtocol_TCP)

type DatastoreArmyKurtosisModule struct {
	numDatstoresAdded int
}

func NewDatastoreArmyKurtosisModule() *DatastoreArmyKurtosisModule {
	return &DatastoreArmyKurtosisModule{}
}

func (module *DatastoreArmyKurtosisModule) Execute(enclaveCtx *enclaves.EnclaveContext, serializedParams string) (serializedResult string, resultError error) {
	params := new(ExecuteParams)
	if err := json.Unmarshal([]byte(serializedParams), params); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred unmarshalling the params JSON")
	}

	createdServiceIdsToPortIds := map[string]string{}
	for i := uint32(0); i < params.NumDatastores; i++ {
		serviceId, err := module.addDatastoreService(enclaveCtx)
		if err != nil {
			return "", stacktrace.Propagate(err, "An error occurred adding a datastore service")
		}
		createdServiceIdsToPortIds[string(serviceId)] = datastorePortId
	}
	resultObj := ExecuteResult{
		CreatedServiceIdsToPortIds: createdServiceIdsToPortIds,
	}
	resultJsonBytes, err := json.Marshal(resultObj)
	if err != nil {
		return "", stacktrace.Propagate(err, "An error occurred serialzing the Lambda result object to JSON")
	}
	return string(resultJsonBytes), nil
}

// ====================================================================================================
//                                       Private helper functions
// ====================================================================================================
func (module *DatastoreArmyKurtosisModule) addDatastoreService(enclaveCtx *enclaves.EnclaveContext) (services.ServiceID, error) {
	nextDatastoreServiceId := services.ServiceID(fmt.Sprintf("datastore-%v", module.numDatstoresAdded))

	datastoreContainerConfigSupplier := getDatastoreContainerConfigSupplier()

	if _, err := enclaveCtx.AddService(nextDatastoreServiceId, datastoreContainerConfigSupplier); err != nil {
		return "", stacktrace.Propagate(err, "An error occurred adding datastore service '%v'", nextDatastoreServiceId)
	}
	module.numDatstoresAdded = module.numDatstoresAdded + 1
	return nextDatastoreServiceId, nil
}

func getDatastoreContainerConfigSupplier() func(ipAddr string) (*services.ContainerConfig, error) {
	containerConfigSupplier := func(ipAddr string) (*services.ContainerConfig, error) {
		containerConfig := services.NewContainerConfigBuilder(
			datastoreImage,
		).WithUsedPorts(map[string]*services.PortSpec{
			datastorePortId: datastorePortSpec,
		}).Build()
		return containerConfig, nil
	}
	return containerConfigSupplier
}
