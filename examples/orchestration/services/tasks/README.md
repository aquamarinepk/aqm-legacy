# Tasks Service (Draft)

Role: CRUD API for collaborative todo items; validates ownership with the Accounts service and emits events for Activity.

Implementation notes:
- Built with the same domain logic showcased in `examples/monolith`, but persistence goes through `aqm.MongoRepo` (Mongo collection `tasks`).
- HTTP exposure relies on the shared middleware stack from `pkg/shared/runtime` + `aqm.WithHTTPMiddleware`.
- Service wiring happens in `main.go` (config, logger, Mongo client, lifecycle/shutdown).
- Container image built via `services/tasks/Dockerfile` and referenced by `deploy/local/docker-compose.yml`.

Pending tasks:
- Define inter-service contracts (DTOs, errors) in `pkg/shared` for Accounts/Activity calls.
- Add outbound hooks (events/webhooks) once the Activity service exists.
- Expand config docs (README) so contributors know how to run the service standalone or via docker-compose.
