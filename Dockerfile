 ## Stage :: build
FROM golang:1.14 AS Builder

# copy the code from the host
WORKDIR $GOPATH/src/github.com/actumn/searchgoose
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix nocgo -o /app .

## Stage :: run
# run
FROM ubuntu
COPY --from=builder /app ./
COPY ./searchgoose.yaml ./
CMD ["./app"]
EXPOSE 8080