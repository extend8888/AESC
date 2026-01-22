#!/bin/bash

jq '.validators = []' ~/.aesc/config/genesis.json > ~/.aesc/config/tmp_genesis.json
cd build/generated/gentx
IDX=0
for FILE in *
do
    jq '.validators['$IDX'] |= .+ {}' ~/.aesc/config/tmp_genesis.json > ~/.aesc/config/tmp_genesis_step_1.json && rm ~/.aesc/config/tmp_genesis.json
    KEY=$(jq '.body.messages[0].pubkey.key' $FILE -c)
    DELEGATION=$(jq -r '.body.messages[0].value.amount' $FILE)
    POWER=$(($DELEGATION / 1000000))
    jq '.validators['$IDX'] += {"power":"'$POWER'"}' ~/.aesc/config/tmp_genesis_step_1.json > ~/.aesc/config/tmp_genesis_step_2.json && rm ~/.aesc/config/tmp_genesis_step_1.json
    jq '.validators['$IDX'] += {"pub_key":{"type":"tendermint/PubKeyEd25519","value":'$KEY'}}' ~/.aesc/config/tmp_genesis_step_2.json > ~/.aesc/config/tmp_genesis_step_3.json && rm ~/.aesc/config/tmp_genesis_step_2.json
    mv ~/.aesc/config/tmp_genesis_step_3.json ~/.aesc/config/tmp_genesis.json
    IDX=$(($IDX+1))
done

mv ~/.aesc/config/tmp_genesis.json ~/.aesc/config/genesis.json
