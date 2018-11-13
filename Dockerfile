FROM quay.io/prometheus/busybox:latest

COPY bin/emq_exporter /bin/emq_exporter

EXPOSE 9505

ENTRYPOINT [ "/bin/emq_exporter" ]