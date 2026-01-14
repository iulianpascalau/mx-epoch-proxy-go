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

