
# Kubernetes config

This directory contains Kubernetes manifest templates for rolling out the PDC Agent.

It contains two manifests: `agent-bare.yaml`, which describes te agent Deployment, and `agent-secret-bare.yaml`, which describes the Secret which the Agent will need to connect to the PDC gateway. Both of these manifests are templates - they contain variables, so they cannot be used as-is.

## Installing 

### 1. Installing the Secret
