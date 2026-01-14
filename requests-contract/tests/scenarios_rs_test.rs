use multiversx_sc_scenario::*;

fn world() -> ScenarioWorld {
    let mut blockchain = ScenarioWorld::new();

    blockchain.set_current_dir_from_workspace("");
    blockchain.register_contract("mxsc:output/requests.mxsc.json", requests::ContractBuilder);
    blockchain
}

#[test]
fn add_requests_accumulation_rs() {
    world().run("scenarios/add_requests_accumulation.scen.json");
}

#[test]
fn add_requests_multiple_rs() {
    world().run("scenarios/add_requests_multiple.scen.json");
}

#[test]
fn add_requests_single_rs() {
    world().run("scenarios/add_requests_single.scen.json");
}

#[test]
fn add_requests_when_paused_rs() {
    world().run("scenarios/add_requests_when_paused.scen.json");
}

#[test]
fn change_rate_non_owner_rs() {
    world().run("scenarios/change_rate_non_owner.scen.json");
}

#[test]
fn change_rate_valid_rs() {
    world().run("scenarios/change_rate_valid.scen.json");
}

#[test]
fn change_rate_zero_rs() {
    world().run("scenarios/change_rate_zero.scen.json");
}

#[test]
fn full_workflow_rs() {
    world().run("scenarios/full_workflow.scen.json");
}

#[test]
fn get_requests_existing_rs() {
    world().run("scenarios/get_requests_existing.scen.json");
}

#[test]
fn get_requests_nonexistent_rs() {
    world().run("scenarios/get_requests_nonexistent.scen.json");
}

#[test]
fn init_valid_rs() {
    world().run("scenarios/init_valid.scen.json");
}

#[test]
fn init_zero_rs() {
    world().run("scenarios/init_zero.scen.json");
}

#[test]
fn pause_already_paused_rs() {
    world().run("scenarios/pause_already_paused.scen.json");
}

#[test]
fn pause_non_owner_rs() {
    world().run("scenarios/pause_non_owner.scen.json");
}

#[test]
fn pause_success_rs() {
    world().run("scenarios/pause_success.scen.json");
}

#[test]
fn pause_unpause_workflow_rs() {
    world().run("scenarios/pause_unpause_workflow.scen.json");
}

#[test]
fn rate_change_affects_future_rs() {
    world().run("scenarios/rate_change_affects_future.scen.json");
}

#[test]
fn unpause_non_owner_rs() {
    world().run("scenarios/unpause_non_owner.scen.json");
}

#[test]
fn unpause_not_paused_rs() {
    world().run("scenarios/unpause_not_paused.scen.json");
}

#[test]
fn unpause_success_rs() {
    world().run("scenarios/unpause_success.scen.json");
}

#[test]
fn withdraw_empty_rs() {
    world().run("scenarios/withdraw_empty.scen.json");
}

#[test]
fn withdraw_non_owner_rs() {
    world().run("scenarios/withdraw_non_owner.scen.json");
}

#[test]
fn withdraw_success_rs() {
    world().run("scenarios/withdraw_success.scen.json");
}
