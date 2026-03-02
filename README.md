# 🚢 Raft3D

### Distributed 3D Printing Management System using RAFT Consensus Algorithm

------------------------------------------------------------------------

## 🌟 Project Overview

**Raft3D** is a fault-tolerant distributed system that manages **3D
printers, filaments, and print jobs** using the **RAFT Consensus
Algorithm**.

The system guarantees **data consistency, reliability, and high
availability** across multiple nodes by replicating application state
within a distributed cluster. Even during node or leader failure, the
system continues operation through automatic **leader election and log
replication**.

------------------------------------------------------------------------

## 🚀 Key Highlights

-   Real implementation of **RAFT Consensus Algorithm**
-   Distributed cluster with Leader--Follower architecture
-   Automatic leader election & recovery
-   Replicated State Machine (FSM)
-   Fault-tolerant system design
-   REST-based distributed API communication
-   Dynamic node joining & removal

------------------------------------------------------------------------

## 🧠 Problem Statement

Centralized systems suffer from: - Single point of failure - Data
inconsistency - Service downtime

**Raft3D solves this** by enabling multiple distributed nodes to agree
on a single consistent system state using consensus.

------------------------------------------------------------------------

## 🏗️ System Architecture

Client Request\
↓\
Leader Node\
↓\
Log Replication\
↓\
Follower Nodes\
↓\
State Synchronization

------------------------------------------------------------------------

## ⚙️ Tech Stack

  Category        Technology
  --------------- ---------------------
  Language        Go (Golang)
  Consensus       HashiCorp RAFT
  API Framework   Gorilla Mux
  Communication   TCP Transport
  Testing         Postman
  System Type     Distributed Systems

------------------------------------------------------------------------

## 📂 Project Structure

    Raft3D/
    │
    ├── main.go          # Node initialization
    ├── raft_node.go     # RAFT configuration
    ├── fsm.go           # Replicated state machine
    ├── server.go        # REST APIs
    └── raft-data/       # Persistent logs

------------------------------------------------------------------------

## 🔄 RAFT Workflow

### 1️⃣ Leader Election

-   Cluster bootstraps with one leader.
-   On leader failure → election triggered.
-   Majority voting selects new leader.

### 2️⃣ Log Replication

-   Leader receives client request.
-   Entry appended to RAFT log.
-   Followers replicate entry.
-   Majority confirmation commits state.

### 3️⃣ Fault Tolerance

System continues operation even after leader crash.

------------------------------------------------------------------------

## 📡 REST API Endpoints

### Printer Management

    POST /api/v1/printers
    GET  /api/v1/printers

### Filament Management

    POST /api/v1/filaments
    GET  /api/v1/filaments

### Print Job Management

    POST /api/v1/print_jobs
    POST /api/v1/print_jobs/{id}/status
    GET  /api/v1/print_jobs

### Cluster Control

    POST /join
    POST /remove

------------------------------------------------------------------------

## 🧪 Failure Handling Demonstration

-   Leader node termination
-   Automatic re-election
-   Continued write operations
-   Consistent replicated state

------------------------------------------------------------------------

## ▶️ Running the Project

### Start Leader

``` bash
go run main.go --id=node1 --http=8000 --raftport=5001 --bootstrap
```

### Start Followers

``` bash
go run main.go --id=node2 --http=8001 --raftport=5002
go run main.go --id=node3 --http=8002 --raftport=5003
```

### Join Cluster

    POST http://localhost:8000/join

------------------------------------------------------------------------

## 📈 Learning Outcomes

-   Distributed System Design
-   Consensus Algorithms
-   Cloud Computing Concepts
-   State Machine Replication
-   Fault Tolerant Architecture
-   Leader-Based Coordination

------------------------------------------------------------------------

## 💡 Real-World Applications

-   Distributed Databases
-   Cloud Infrastructure
-   Kubernetes Control Systems
-   Microservice Coordination
-   IoT Device Clusters
-   Industrial Automation Networks

------------------------------------------------------------------------

## 👨‍💻 Team

-   Sai Mourya N Doddamani\
-   Sai Ganesh\
-   Samarth NN\
-   **Shashank R Patil**\
-   Vignesh

**PES University --- Cloud Computing Mini Project**


