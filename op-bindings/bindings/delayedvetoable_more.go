// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package bindings

import (
	"encoding/json"

	"github.com/ethereum-optimism/optimism/op-bindings/solc"
)

const DelayedVetoableStorageLayoutJSON = "{\"storage\":[{\"astId\":1000,\"contract\":\"src/L1/DelayedVetoable.sol:DelayedVetoable\",\"label\":\"_delay\",\"offset\":0,\"slot\":\"0\",\"type\":\"t_uint256\"},{\"astId\":1001,\"contract\":\"src/L1/DelayedVetoable.sol:DelayedVetoable\",\"label\":\"_queuedAt\",\"offset\":0,\"slot\":\"1\",\"type\":\"t_mapping(t_bytes32,t_uint256)\"}],\"types\":{\"t_bytes32\":{\"encoding\":\"inplace\",\"label\":\"bytes32\",\"numberOfBytes\":\"32\"},\"t_mapping(t_bytes32,t_uint256)\":{\"encoding\":\"mapping\",\"label\":\"mapping(bytes32 =\u003e uint256)\",\"numberOfBytes\":\"32\",\"key\":\"t_bytes32\",\"value\":\"t_uint256\"},\"t_uint256\":{\"encoding\":\"inplace\",\"label\":\"uint256\",\"numberOfBytes\":\"32\"}}}"

var DelayedVetoableStorageLayout = new(solc.StorageLayout)

var DelayedVetoableDeployedBin = "0x608060405234801561001057600080fd5b50600436106100725760003560e01c8063b912de5d11610050578063b912de5d14610111578063d4b8399214610124578063d8bff4401461012c57610072565b806354fd4d501461007c5780635c39fcc1146100ce5780636a42b8f8146100fb575b61007a610134565b005b6100b86040518060400160405280600581526020017f312e302e3000000000000000000000000000000000000000000000000000000081525081565b6040516100c591906106a7565b60405180910390f35b6100d66104fb565b60405173ffffffffffffffffffffffffffffffffffffffff90911681526020016100c5565b610103610532565b6040519081526020016100c5565b61010361011f36600461071a565b610540565b6100d6610567565b6100d6610593565b361580156101425750600054155b15610298573373ffffffffffffffffffffffffffffffffffffffff7f000000000000000000000000000000000000000000000000000000000000000016148015906101c357503373ffffffffffffffffffffffffffffffffffffffff7f00000000000000000000000000000000000000000000000000000000000000001614155b1561023d576040517f295a81c100000000000000000000000000000000000000000000000000000000815273ffffffffffffffffffffffffffffffffffffffff7f00000000000000000000000000000000000000000000000000000000000000001660048201523360248201526044015b60405180910390fd5b7f000000000000000000000000000000000000000000000000000000000000000060008190556040519081527febf28bfb587e28dfffd9173cf71c32ba5d3f0544a0117b5539c9b274a5bba2a89060200160405180910390a1565b600080366040516102aa929190610733565b60405190819003902090503373ffffffffffffffffffffffffffffffffffffffff7f0000000000000000000000000000000000000000000000000000000000000000161480156103065750600081815260016020526040902054155b1561036c5760005460000361031e5761031e816105bf565b6000818152600160205260408082204290555182917f87a332a414acbc7da074543639ce7ae02ff1ea72e88379da9f261b080beb5a139161036191903690610743565b60405180910390a250565b3373ffffffffffffffffffffffffffffffffffffffff7f0000000000000000000000000000000000000000000000000000000000000000161480156103be575060008181526001602052604090205415155b15610406576000818152600160205260408082208290555182917fbede6852c1d97d93ff557f676de76670cd0dec861e7fe8beb13aa0ba2b0ab0409161036191903690610743565b600081815260016020526040812054900361048b576040517f295a81c100000000000000000000000000000000000000000000000000000000815273ffffffffffffffffffffffffffffffffffffffff7f0000000000000000000000000000000000000000000000000000000000000000166004820152336024820152604401610234565b60008054828252600160205260409091205442916104a891610790565b10156104e0576040517f43dc986d00000000000000000000000000000000000000000000000000000000815260040160405180910390fd5b6000818152600160205260408120556104f8816105bf565b50565b60003361052757507f000000000000000000000000000000000000000000000000000000000000000090565b61052f610134565b90565b600033610527575060005490565b60003361055a575060009081526001602052604090205490565b610562610134565b919050565b60003361052757507f000000000000000000000000000000000000000000000000000000000000000090565b60003361052757507f000000000000000000000000000000000000000000000000000000000000000090565b807f4c109d85bcd0bb5c735b4be850953d652afe4cd9aa2e0b1426a65a4dcb2e12296000366040516105f2929190610743565b60405180910390a26000807f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff16600036604051610645929190610733565b6000604051808303816000865af19150503d8060008114610682576040519150601f19603f3d011682016040523d82523d6000602084013e610687565b606091505b50909250905081151560010361069f57805160208201f35b805160208201fd5b600060208083528351808285015260005b818110156106d4578581018301518582016040015282016106b8565b818111156106e6576000604083870101525b50601f017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe016929092016040019392505050565b60006020828403121561072c57600080fd5b5035919050565b8183823760009101908152919050565b60208152816020820152818360408301376000818301604090810191909152601f9092017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0160101919050565b600082198211156107ca577f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b50019056fea164736f6c634300080f000a"


var DelayedVetoableImmutableReferencesJSON = "{\"74966\":[{\"start\":1393,\"length\":32},{\"start\":1535,\"length\":32}],\"74969\":[{\"start\":416,\"length\":32},{\"start\":900,\"length\":32},{\"start\":1437,\"length\":32}],\"74972\":[{\"start\":351,\"length\":32},{\"start\":517,\"length\":32},{\"start\":717,\"length\":32},{\"start\":1112,\"length\":32},{\"start\":1285,\"length\":32}],\"74975\":[{\"start\":575,\"length\":32}]}"

func init() {
	if err := json.Unmarshal([]byte(DelayedVetoableStorageLayoutJSON), DelayedVetoableStorageLayout); err != nil {
		panic(err)
	}

	layouts["DelayedVetoable"] = DelayedVetoableStorageLayout
	deployedBytecodes["DelayedVetoable"] = DelayedVetoableDeployedBin
	immutableReferences["DelayedVetoable"] = DelayedVetoableImmutableReferencesJSON
}
