FROM progrium/busybox
MAINTAINER Jeffrey Jenner <thetooth@ameoto.com>

COPY ./static /opt/server/static
COPY ./template.html /opt/server/template.html
COPY ./main /opt/server/main
RUN chmod 755 /opt/server/main
WORKDIR "/opt/server"

CMD ["./main"]
