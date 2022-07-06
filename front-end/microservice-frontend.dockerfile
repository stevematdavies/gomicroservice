FROM alpine:latest
RUN mkdir /app
COPY frontEndApp /app
COPY templates /templates
CMD ["/app/frontEndApp"]