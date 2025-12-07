// SPDX-License-Identifier: MIT
pragma solidity ^0.8.25;

import {ContractRegistry} from "@flarenetwork/flare-periphery-contracts/coston2/ContractRegistry.sol";
import {TestFtsoV2Interface} from "@flarenetwork/flare-periphery-contracts/coston2/TestFtsoV2Interface.sol";
import {IEVMTransaction} from "@flarenetwork/flare-periphery-contracts/coston2/IEVMTransaction.sol";
import {IFdcVerification} from "@flarenetwork/flare-periphery-contracts/coston2/IFdcVerification.sol";

contract FlareFTSOFDCExample {
    // FTSO feed ID for FLR/USD on Coston2
    bytes21 public constant FLR_USD_ID = 
        0x01464c522f55534400000000000000000000000000; // "FLR/USD"
    
    // Sepolia USDC contract for event filtering
    address public constant SEPOLIA_USDC = 
        0x1c7D4B196Cb0C7B01d743Fbc6116a902379C7238;
    
    // Transfer events storage
    struct VerifiedTransfer {
        address from;
        address to;
        uint256 amount;
        uint64 timestamp;
    }
    
    VerifiedTransfer[] public verifiedTransfers;
    uint256 public totalDeposited;
    
    // Events
    event PriceChecked(uint256 flrUsdPrice, int8 decimals);
    event TransferVerified(address from, address to, uint256 amount);
    event CrossChainDeposit(address user, uint256 amount);

    // === FTSO INTEGRATION ===
    function getFlrUsdPrice() public view returns (uint256 value, int8 decimals, uint64 timestamp) {
        // Fetch FTSOv2 via registry (test interface for view calls, no gas in Remix)
        TestFtsoV2Interface ftsoV2 = ContractRegistry.getTestFtsoV2();
        return ftsoV2.getFeedById(FLR_USD_ID);
    }
    
    // === FDC INTEGRATION ===
    function verifyAndStoreTransfer(IEVMTransaction.Proof calldata _proof) public {
        // 1. Verify via FDC (checks Merkle proof against onchain consensus root)
        IFdcVerification fdc = ContractRegistry.getFdcVerification();
        require(fdc.verifyEVMTransaction(_proof), "Invalid FDC proof");
        
        // 2. Parse events from verified tx
        IEVMTransaction.ResponseBody memory resp = _proof.data.responseBody;
        for (uint256 i = 0; i < resp.events.length; i++) {
            IEVMTransaction.Event memory evt = resp.events[i];
            
            // Filter: USDC Transfer events only
            if (evt.emitterAddress == SEPOLIA_USDC &&
                evt.topics.length >= 3 &&
                evt.topics[0] == 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef) { // Transfer(address,address,uint256) topic0
                
                // Decode: topics[1]=from, topics[2]=to, data=amount
                address from = address(uint160(uint256(evt.topics[1])));
                address to = address(uint160(uint256(evt.topics[2])));
                uint256 amount = abi.decode(evt.data, (uint256));
                
                verifiedTransfers.push(VerifiedTransfer(from, to, amount, resp.timestamp));
                emit TransferVerified(from, to, amount);
            }
        }
    }
    
    // === BUSINESS LOGIC: Price-aware cross-chain deposit ===
    function crossChainDeposit(IEVMTransaction.Proof calldata _proof) external {
        // Verify FDC proof first
        verifyAndStoreTransfer(_proof);
        
        // Get latest FLR/USD price via FTSO
        (uint256 flrPrice,,) = getFlrUsdPrice();
        emit PriceChecked(flrPrice, 5); // Assume 5 decimals for demo
        
        // Example: Only allow if latest transfer > $10 FLR equivalent
        VerifiedTransfer memory latest = verifiedTransfers[verifiedTransfers.length - 1];
        require(latest.amount >= flrPrice * 10, "Transfer below FLR price threshold");
        
        // "Deposit" logic (in real app: mint tokens, etc.)
        totalDeposited += latest.amount;
        emit CrossChainDeposit(msg.sender, latest.amount);
    }
    
    // View helpers
    function getLatestTransfer() external view returns (VerifiedTransfer memory) {
        require(verifiedTransfers.length > 0, "No transfers");
        return verifiedTransfers[verifiedTransfers.length - 1];
    }
    
    function getTotalVerifiedTransfers() external view returns (uint256) {
        return verifiedTransfers.length;
    }
}
