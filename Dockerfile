ARG ARG_BUILDNUMBER=none

# build the frontend
FROM node:16.18.0-alpine3.16 AS frontendbuilder

WORKDIR /app

COPY frontend/package.json ./
COPY frontend/package-lock.json ./

RUN npm install

COPY frontend/ .

RUN npm run build

# build backend
FROM golang:1.23-alpine AS build

RUN apk add --no-cache git

WORKDIR /src

COPY ./ ./

# Run tests​
# RUN CGO_ENABLED=0 go test -timeout 30s -v ./...

COPY --from=frontendbuilder /app/build/ /src/backend/cmd/nomad-ops-server/wwwroot/

# Build the executable​
RUN CGO_ENABLED=0 go build \
    -mod=vendor \
    -o /app ./backend/cmd/nomad-ops-server/

# STAGE 2: build the container to run​
FROM gcr.io/distroless/static AS final

ENV BUILDNUMBER=$ARG_BUILDNUMBER

#USER nonroot:nonroot


# copy compiled app​
COPY --from=build /app /app

ENTRYPOINT ["/app"]
