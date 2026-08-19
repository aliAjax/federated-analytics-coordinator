# Architecture, Security and Operations

## Domain model

A participant is tenant-scoped, has one active Ed25519 public key and declares
metric capabilities. A task owns its participant set, deadline, state machine,
privacy budget and revision. A share has a task/participant/round/nonce tuple
and a signature over its canonical representation. A released result carries
only the aggregate, count, epsilon and precision statement. Audits are
append-only hash chains.

## State machine

`draft -> collecting -> masked -> aggregating -> revealing -> completed`; an
operator or deadline handler may move any nonterminal state to `aborted`.
Budget is reserved on collection, committed when result release succeeds, and
released when aborting. A distributed implementation must fence state updates
on the revision and store each transition with an outbox entry.

## Threat model

The adversaries are a replaying client, a compromised participant key, a
malicious share contributor, an operator reading data, and an unavailable
participant. The implementation blocks nonce replay, validates signatures,
thresholds participants, checks fixed-point overflow, and does not expose
individual shares through public result routes. Production must use mTLS,
sealed secret storage, key revocation, HSM-backed signing where appropriate,
encrypted payload storage and monitoring that avoids raw numeric values.

## Capacity and runbook

Share validation is O(number of shares) per aggregation and constant memory per
request. At 10,000 concurrent tasks with 20 participants, durable shares
should be partitioned by task id and workers lease one partition at a time.
Target SLO: 99.9% share acceptance below 100ms excluding client network time.
For incident drills: expire a lease while aggregating, retry a nonce, submit an
invalid signature, abort before release, and verify the audit hash chain and
budget release. Capture pprof only while authorized for an incident.
