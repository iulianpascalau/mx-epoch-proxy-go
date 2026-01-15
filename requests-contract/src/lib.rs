#![no_std]

use multiversx_sc::imports::*;

#[multiversx_sc::contract]
pub trait RequestsContract {
    /// Constructor - initializes the contract with the number of requests per EGLD
    #[init]
    fn init(&self, num_requests_per_egld: BigUint) {
        require!(num_requests_per_egld > 0, "Number of requests per EGLD must be non-zero");
        self.num_requests_per_egld().set(num_requests_per_egld);
        self.is_paused().set(false);
    }

    /// Upgrade function - called when contract is upgraded
    #[upgrade]
    fn upgrade(&self, num_requests_per_egld: BigUint) {
        require!(num_requests_per_egld > 0, "Number of requests per EGLD must be non-zero");
        self.num_requests_per_egld().set(num_requests_per_egld);
        if !self.is_paused().is_empty() {
            // is_paused already exists, keep current value
        } else {
            // First upgrade, initialize pause state
            self.is_paused().set(false);
        }
    }

    /// Add acquired requests for a given ID - payable only in EGLD
    /// The number of acquired requests added = (EGLD amount transferred in regular units) * num_requests_per_egld
    /// Example: 2.5 EGLD * 100 rate = 250 acquired requests
    #[payable("EGLD")]
    #[endpoint(addRequests)]
    fn add_requests(&self, id: u64) {
        require!(!self.is_paused().get(), "Contract is paused");

        let payment = self.call_value().egld();
        let amount_wei = payment.clone_value();

        require!(amount_wei > 0, "Payment amount must be greater than 0");

        // Convert from wei to EGLD (1 EGLD = 10^18 wei)
        let one_egld = BigUint::from(1_000_000_000_000_000_000u64);

        let num_requests_per_egld = self.num_requests_per_egld().get();
        let requests_to_add = (amount_wei.clone() * &num_requests_per_egld) / one_egld;

        self.acquired_requests(&id).update(|requests| *requests += requests_to_add.clone());

        self.add_requests_event(&id, &amount_wei, &requests_to_add);
    }


    /// Get the number of acquired requests for a given ID
    /// Returns 0 if the ID was not credited
    #[view(getRequests)]
    fn get_requests(&self, id: u64) -> BigUint {
        self.acquired_requests(&id).get()
    }

    /// Check if the contract is paused
    #[view(isPaused)]
    fn get_is_paused(&self) -> bool {
        self.is_paused().get()
    }


    /// Change the number of requests per EGLD
    /// Can only be called by the owner
    #[endpoint(changeNumRequestsPerEGLD)]
    fn change_num_requests_per_egld(&self, new_num_requests_per_egld: BigUint) {
        let caller = self.blockchain().get_caller();
        let owner = self.blockchain().get_owner_address();

        require!(caller == owner, "Only the owner can change the exchange rate");
        require!(new_num_requests_per_egld > 0, "Number of requests per EGLD must be non-zero");

        let old_value = self.num_requests_per_egld().get();
        self.num_requests_per_egld().set(new_num_requests_per_egld.clone());

        self.change_num_requests_per_egld_event(&old_value, &new_num_requests_per_egld);
    }

    /// Pause the contract - prevents new requests from being added
    /// Can only be called by the owner
    #[endpoint(pause)]
    fn pause(&self) {
        let caller = self.blockchain().get_caller();
        let owner = self.blockchain().get_owner_address();

        require!(caller == owner, "Only the owner can pause the contract");
        require!(!self.is_paused().get(), "Contract is already paused");

        self.is_paused().set(true);
        self.pause_event();
    }

    /// Unpause the contract - allows new requests to be added again
    /// Can only be called by the owner
    #[endpoint(unpause)]
    fn unpause(&self) {
        let caller = self.blockchain().get_caller();
        let owner = self.blockchain().get_owner_address();

        require!(caller == owner, "Only the owner can unpause the contract");
        require!(self.is_paused().get(), "Contract is not paused");

        self.is_paused().set(false);
        self.unpause_event();
    }

    /// Withdraw all available EGLD in the contract to the owner's address
    /// Can only be called by the owner
    #[endpoint(withdrawAll)]
    fn withdraw_all(&self) {
        let caller = self.blockchain().get_caller();
        let owner = self.blockchain().get_owner_address();

        require!(caller == owner, "Only the owner can withdraw");

        let contract_balance = self.blockchain().get_sc_balance(&EgldOrEsdtTokenIdentifier::egld(), 0);
        require!(contract_balance > 0, "No EGLD to withdraw");

        self.tx()
            .to(&owner)
            .egld(&contract_balance)
            .transfer();

        self.withdraw_event(&owner, &contract_balance);
    }

    /// Event emitted when requests are added
    #[event("addRequests")]
    fn add_requests_event(
        &self,
        #[indexed] id: &u64,
        #[indexed] egld_amount: &BigUint,
        requests_added: &BigUint,
    );

    /// Event emitted when the exchange rate is changed
    #[event("changeNumRequestsPerEGLD")]
    fn change_num_requests_per_egld_event(
        &self,
        #[indexed] old_value: &BigUint,
        new_value: &BigUint,
    );

    /// Event emitted when the contract is paused
    #[event("pause")]
    fn pause_event(&self);

    /// Event emitted when the contract is unpaused
    #[event("unpause")]
    fn unpause_event(&self);

    /// Event emitted when EGLD is withdrawn
    #[event("withdraw")]
    fn withdraw_event(
        &self,
        #[indexed] recipient: &ManagedAddress,
        amount: &BigUint,
    );

    /// Storage mapper for the number of requests per EGLD
    #[storage_mapper("numRequestsPerEgld")]
    fn num_requests_per_egld(&self) -> SingleValueMapper<BigUint>;

    /// Storage mapper for acquired requests count per ID
    #[storage_mapper("acquiredRequests")]
    fn acquired_requests(&self, id: &u64) -> SingleValueMapper<BigUint>;

    /// Storage mapper for pause state
    #[storage_mapper("isPaused")]
    fn is_paused(&self) -> SingleValueMapper<bool>;
}
