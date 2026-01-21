.PHONY: migrate-up migrate-down migrate-status migrate-create migrate-redo migrate-reset

migrate-up:
	goose up

migrate-down:
	goose down

migrate-status:
	goose status

migrate-redo:
	goose redo

migrate-reset:
	goose reset

migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "Usage: make migrate-create name=add_users"; \
		exit 1; \
	fi
	goose create $(name) sql

