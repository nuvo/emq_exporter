FROM quay.io/prometheus/busybox:latest

COPY emq_exporter /bin/emq_exporter

EXPOSE 9505
USER nobody

ENTRYPOINT [ "/bin/emq_exporter" ]