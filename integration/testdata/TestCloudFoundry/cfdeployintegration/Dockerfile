FROM maven:3.6.3-openjdk-8-slim

RUN  apt-get -y update; apt-get -y install gnupg wget
RUN wget -q -O - https://packages.cloudfoundry.org/debian/cli.cloudfoundry.org.key |  apt-key add -
RUN echo "deb https://packages.cloudfoundry.org/debian stable main" |  tee /etc/apt/sources.list.d/cloudfoundry-cli.list
RUN  apt-get -y update; apt-get -y install cf-cli
COPY run-in-container.sh /test.sh
