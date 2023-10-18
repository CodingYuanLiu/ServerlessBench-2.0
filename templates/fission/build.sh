#!/bin/sh
pip3 install -r ${SRC_PKG}/requirements.txt -t ${SRC_PKG} && cp -r -f ${SRC_PKG} ${DEPLOY_PKG}