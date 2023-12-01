// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package bindings

import (
	"encoding/json"

	"github.com/ethereum-optimism/optimism/op-bindings/solc"
)

const SequencerFeeVaultStorageLayoutJSON = "{\"storage\":[{\"astId\":1000,\"contract\":\"src/L2/SequencerFeeVault.sol:SequencerFeeVault\",\"label\":\"totalProcessed\",\"offset\":0,\"slot\":\"0\",\"type\":\"t_uint256\"}],\"types\":{\"t_uint256\":{\"encoding\":\"inplace\",\"label\":\"uint256\",\"numberOfBytes\":\"32\"}}}"

var SequencerFeeVaultStorageLayout = new(solc.StorageLayout)

var SequencerFeeVaultDeployedBin = "0x6080604052600436106100745760003560e01c806384411d651161004e57806384411d651461014b578063d0e12f901461016f578063d3e5792b146101b0578063d4ff9218146101e457600080fd5b80630d9019e1146100805780633ccfd60b146100de57806354fd4d50146100f557600080fd5b3661007b57005b600080fd5b34801561008c57600080fd5b506100b47f000000000000000000000000000000000000000000000000000000000000000081565b60405173ffffffffffffffffffffffffffffffffffffffff90911681526020015b60405180910390f35b3480156100ea57600080fd5b506100f3610217565b005b34801561010157600080fd5b5061013e6040518060400160405280600581526020017f312e342e3100000000000000000000000000000000000000000000000000000081525081565b6040516100d5919061066e565b34801561015757600080fd5b5061016160005481565b6040519081526020016100d5565b34801561017b57600080fd5b506101a37f000000000000000000000000000000000000000000000000000000000000000081565b6040516100d591906106f2565b3480156101bc57600080fd5b506101617f000000000000000000000000000000000000000000000000000000000000000081565b3480156101f057600080fd5b507f00000000000000000000000000000000000000000000000000000000000000006100b4565b7f00000000000000000000000000000000000000000000000000000000000000004710156102f2576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152604a60248201527f4665655661756c743a207769746864726177616c20616d6f756e74206d75737460448201527f2062652067726561746572207468616e206d696e696d756d207769746864726160648201527f77616c20616d6f756e7400000000000000000000000000000000000000000000608482015260a4015b60405180910390fd5b6000479050806000808282546103089190610706565b9091555050604080518281527f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff166020820152338183015290517fc8a211cc64b6ed1b50595a9fcb1932b6d1e5a6e8ef15b60e5b1f988ea9086bba9181900360600190a17f38e04cbeb8c10f8f568618aa75be0f10b6729b8b4237743b4de20cbcde2839ee817f0000000000000000000000000000000000000000000000000000000000000000337f00000000000000000000000000000000000000000000000000000000000000006040516103f69493929190610745565b60405180910390a160017f0000000000000000000000000000000000000000000000000000000000000000600181111561043257610432610688565b0361054b5760007f000000000000000000000000000000000000000000000000000000000000000073ffffffffffffffffffffffffffffffffffffffff168260405160006040518083038185875af1925050503d80600081146104b1576040519150601f19603f3d011682016040523d82523d6000602084013e6104b6565b606091505b5050905080610547576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152603060248201527f4665655661756c743a206661696c656420746f2073656e642045544820746f2060448201527f4c322066656520726563697069656e740000000000000000000000000000000060648201526084016102e9565b5050565b604080516020810182526000815290517fe11013dd0000000000000000000000000000000000000000000000000000000081527342000000000000000000000000000000000000109163e11013dd9184916105ce917f0000000000000000000000000000000000000000000000000000000000000000916188b891600401610786565b6000604051808303818588803b1580156105e757600080fd5b505af11580156105fb573d6000803e3d6000fd5b505050505050565b6000815180845260005b818110156106295760208185018101518683018201520161060d565b8181111561063b576000602083870101525b50601f017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0169290920160200192915050565b6020815260006106816020830184610603565b9392505050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052602160045260246000fd5b600281106106ee577f4e487b7100000000000000000000000000000000000000000000000000000000600052602160045260246000fd5b9052565b6020810161070082846106b7565b92915050565b60008219821115610740577f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b500190565b84815273ffffffffffffffffffffffffffffffffffffffff8481166020830152831660408201526080810161077d60608301846106b7565b95945050505050565b73ffffffffffffffffffffffffffffffffffffffff8416815263ffffffff8316602082015260606040820152600061077d606083018461060356fea164736f6c634300080f000a"


var SequencerFeeVaultImmutableReferencesJSON = "{\"95011\":[{\"start\":450,\"length\":32},{\"start\":537,\"length\":32}],\"95014\":[{\"start\":146,\"length\":32},{\"start\":499,\"length\":32},{\"start\":790,\"length\":32},{\"start\":933,\"length\":32},{\"start\":1083,\"length\":32},{\"start\":1442,\"length\":32}],\"95018\":[{\"start\":385,\"length\":32},{\"start\":967,\"length\":32},{\"start\":1026,\"length\":32}]}"

func init() {
	if err := json.Unmarshal([]byte(SequencerFeeVaultStorageLayoutJSON), SequencerFeeVaultStorageLayout); err != nil {
		panic(err)
	}

	layouts["SequencerFeeVault"] = SequencerFeeVaultStorageLayout
	deployedBytecodes["SequencerFeeVault"] = SequencerFeeVaultDeployedBin
	immutableReferences["SequencerFeeVault"] = SequencerFeeVaultImmutableReferencesJSON
}
