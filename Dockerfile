FROM alpine
EXPOSE 9200
COPY config.json /
COPY jmx_json_exporter /
ENTRYPOINT ["/jmx_json_exporter"]
