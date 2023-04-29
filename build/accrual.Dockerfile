# syntax=docker/dockerfile:1

# Deploy the application binary into a lean image
FROM gcr.io/distroless/base-debian11 AS build-release-stage

WORKDIR /app

COPY ./cmd/accrual/accrual_linux_amd64 /accrual

EXPOSE $RUN_PORT

USER nonroot:nonroot

CMD ["/accrual"]