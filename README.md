# Aquamarine

**Another generator? From the same author?**
Yes. Aquamarine started as a Go generator for monoliths. Early versions used embedding to reduce repetition. Embedding is idiomatic Go, but using it as the default made intent less clear and behavior harder to see. This iteration removes that opacity and adds microservice generation.

## What it is
Aquamarine pairs a small set of building blocks in `aquamarinepk/aqm` with a generator that writes plain Go. `aqm` provides shared concerns like config, logging, metrics, health, and basic observability. In the generated code, usage is explicit: call it directly or wrap it with small adapters. Reuse reads like normal Go.

## CLI
Short commands to bootstrap a project and add entities. Optional per-entity YAML is available when it is more comfortable than long flag lists.

## Status
The public generator at https://github.com/aquamarinepk/aquamarine does not use `aqm` yet. 
Migration is in progress.
