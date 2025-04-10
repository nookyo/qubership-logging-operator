# All releases https://github.com/Netcracker/qubership-docker-integration-tests/releases
# Based on python:3.13.1-alpine3.21
FROM ghcr.io/netcracker/qubership-docker-integration-tests:main

ENV ROBOT_OUTPUT=${ROBOT_HOME}/output \
    ROBOT_UID=1000 \
    ROBOT_GID=0 \
    DISTR_DIR=/tmp/deps \
    SERVICE_CHECKER_SCRIPT=${ROBOT_HOME}/logging_pods_checker.py \
    SERVICE_CHECKER_SCRIPT_TIMEOUT=500

# Copy tests parts
COPY requirements.txt ${ROBOT_HOME}/requirements.txt
COPY src/ ${ROBOT_HOME}/

RUN \
    # Install python requirements
    python3 -m pip install --no-cache-dir -r ${ROBOT_HOME}/requirements.txt \
    # Create and grant permissions for tests output
    && mkdir -p ${ROBOT_OUTPUT} \
    && chmod a+w ${ROBOT_OUTPUT}

USER ${ROBOT_UID}:${ROBOT_GID}

EXPOSE 8080
VOLUME ["${ROBOT_OUTPUT}"]
