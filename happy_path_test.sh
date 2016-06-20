#!/usr/bin/env bash

CMD=./rcm

CL_RST='\033[0m'
CL_RED='\033[0;31m'
CL_GRE='\033[0;32m'
CL_YEL='\033[0;33m'
CL_BLU='\033[0;34m'
CL_CYA='\033[0;36m'

TEST_CMD="${CL_CYA}CMD:${CL_RST}"
TEST_INFO="${CL_BLU}INFO:${CL_RST}"
TEST_SUCCEED="${CL_GRE}Test succeed!${CL_RST}"
TEST_FAILED="${CL_RED}Test failed!${CL_RST}"

ALL_TEST_SUCCEED="${CL_GRE}All tests succeed!${CL_RST}"
ALL_TESTS_FAILED="${CL_RED}All tests failed!${CL_RST}"
SOME_OF_TESTS_FAILED="${CL_YEL}All test failed!${CL_RST}"

TEST_COUNT=0
SUCCEED_TEST_COUNT=0

function random-string {
    openssl rand -hex $1
}

function rcm {
    CMD="./rcm $@ ${CLUSTER_NAME}"

    echo ""
    echo -e "${TEST_CMD} ${CMD}"
    exec ${CMD}
    echo ""
}

function perform-test() {

    ((TEST_COUNT+=1))

    echo -e "${TEST_INFO} Cleaning probable previously created instances" # Just in case

    for pid in $(ps -xo pid,command | grep redis-server | grep cluster | awk '{print $1}'); do
        kill ${pid};
    done;

    CLUSTER_NAME=$(random-string 12)
    NODE_COUNT=$1

    echo -e "${TEST_INFO} Performing test with ${NODE_COUNT} node cluster"

    echo "y" | rcm create -n ${NODE_COUNT}
    echo "" | rcm start
    sleep 2
    echo "y" | rcm distribute-slots
    sleep 3

    echo "y" | rcm damage
    sleep 1

    echo "y" | rcm damage -n "100%"
    sleep 1

    echo "y" | rcm stop
    sleep 1
    echo "y" | rcm remove

    PROCESS_COUNT=$(ps -xo command | grep redis | grep cluster | wc -l)

    if [ "${PROCESS_COUNT}" -gt 0 ]; then
        echo -e ${TEST_FAILED}
        return
    else
        echo -e ${TEST_SUCCEED}
    fi

    ((SUCCEED_TEST_COUNT+=1))
}

set -e

for i in {2..15}; do
    perform-test ${i}
done

if [ "${SUCCEED_TEST_COUNT}" -eq "${TEST_COUNT}" ]; then
    echo -e "${ALL_TEST_SUCCEED}"
elif [ "${SUCCEED_TEST_COUNT}" -eq 0 ]; then
    echo -e "${ALL_TESTS_FAILED}"
else
    echo -e "${SOME_OF_TESTS_FAILED}: ${SUCCEED_TEST_COUNT} succeed out of ${TEST_COUNT}"
fi



