#!/bin/bash

# 完全模仿 docker/localnode/scripts/step3_add_validator_to_genesis.sh
# 只添加 power 和 pub_key，不添加 address 和 name
# collect-gentxs 会自动补充这些字段

jq '.validators = []' ~/.sei/config/genesis.json > ~/.sei/config/tmp_genesis.json
cd build/generated/gentx
IDX=0
for FILE in *
do
    jq '.validators['$IDX'] |= .+ {}' ~/.sei/config/tmp_genesis.json > ~/.sei/config/tmp_genesis_step_1.json && rm ~/.sei/config/tmp_genesis.json
    KEY=$(jq '.body.messages[0].pubkey.key' $FILE -c)
    DELEGATION=$(jq -r '.body.messages[0].value.amount' $FILE)
    # 使用 awk 或 bc 来处理大数字，避免 bash 整数溢出
    # 移除 "uaex" 后缀，然后除以 1000000
    DELEGATION_NUM=${DELEGATION%uaex}
    POWER=$(echo "$DELEGATION_NUM / 1000000" | bc)
    jq '.validators['$IDX'] += {"power":"'$POWER'"}' ~/.sei/config/tmp_genesis_step_1.json > ~/.sei/config/tmp_genesis_step_2.json && rm ~/.sei/config/tmp_genesis_step_1.json
    jq '.validators['$IDX'] += {"pub_key":{"type":"tendermint/PubKeyEd25519","value":'$KEY'}}' ~/.sei/config/tmp_genesis_step_2.json > ~/.sei/config/tmp_genesis_step_3.json && rm ~/.sei/config/tmp_genesis_step_2.json
    mv ~/.sei/config/tmp_genesis_step_3.json ~/.sei/config/tmp_genesis.json
    IDX=$(($IDX+1))
done

mv ~/.sei/config/tmp_genesis.json ~/.sei/config/genesis.json

