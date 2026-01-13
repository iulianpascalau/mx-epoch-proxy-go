use multiversx_sc::types::{Address, BigUint};
use multiversx_sc_scenario::*;

fn world() -> ScenarioWorld {
    let mut blockchain = ScenarioWorld::new();
    blockchain.register_contract(
        "mxsc:output/requests-contract.mxsc.json",
        requests_contract::ContractBuilder,
    );
    blockchain
}

#[test]
fn test_init_with_valid_value() {
    let mut world = world();
    world.run("scenarios/init_valid.scen.json");
}

#[test]
fn test_init_with_zero_value() {
    let mut world = world();
    world.run("scenarios/init_zero.scen.json");
}

#[test]
fn test_add_requests_single_user() {
    let mut world = world();
    world.run("scenarios/add_requests_single.scen.json");
}

#[test]
fn test_add_requests_multiple_users() {
    let mut world = world();
    world.run("scenarios/add_requests_multiple.scen.json");
}

#[test]
fn test_add_requests_accumulation() {
    let mut world = world();
    world.run("scenarios/add_requests_accumulation.scen.json");
}

#[test]
fn test_get_requests_existing_user() {
    let mut world = world();
    world.run("scenarios/get_requests_existing.scen.json");
}

#[test]
fn test_get_requests_nonexistent_user() {
    let mut world = world();
    world.run("scenarios/get_requests_nonexistent.scen.json");
}

#[test]
fn test_change_exchange_rate_valid() {
    let mut world = world();
    world.run("scenarios/change_rate_valid.scen.json");
}

#[test]
fn test_change_exchange_rate_zero() {
    let mut world = world();
    world.run("scenarios/change_rate_zero.scen.json");
}

#[test]
fn test_change_exchange_rate_non_owner() {
    let mut world = world();
    world.run("scenarios/change_rate_non_owner.scen.json");
}

#[test]
fn test_withdraw_all_success() {
    let mut world = world();
    world.run("scenarios/withdraw_success.scen.json");
}

#[test]
fn test_withdraw_all_empty_contract() {
    let mut world = world();
    world.run("scenarios/withdraw_empty.scen.json");
}

#[test]
fn test_withdraw_all_non_owner() {
    let mut world = world();
    world.run("scenarios/withdraw_non_owner.scen.json");
}

#[test]
fn test_full_workflow() {
    let mut world = world();
    world.run("scenarios/full_workflow.scen.json");
}

#[test]
fn test_rate_change_affects_future_requests() {
    let mut world = world();
    world.run("scenarios/rate_change_affects_future.scen.json");
}
