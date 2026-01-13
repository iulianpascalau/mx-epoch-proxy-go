use multiversx_sc_meta::*;

fn main() {
    let mut meta = ScMetaBuilder::default()
        .with_contract_crate_path("../")
        .build();

    meta.compile_and_generate_abi();
}
