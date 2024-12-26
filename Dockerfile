FROM golang:1.23-alpine AS build
WORKDIR /go/src/app
COPY . .
RUN CGO_ENABLED=0 go build -o /go/bin/1px

FROM gcr.io/distroless/static-debian12
WORKDIR /app
COPY --from=build /go/bin/1px .
COPY --from=build /go/src/app/index.tmpl .
EXPOSE 3939
ENTRYPOINT [ "/app/1px" ]
